package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/repositories"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

const (
	inviteCodeLength       = 6
	maxInviteCodeBatchSize = 100
)

var (
	ErrInviteCodeInvalid     = errors.New("invalid invite code")
	ErrInviteCodeRateLimited = errors.New("invite code attempts temporarily limited")
)

var InviteCodeService = newInviteCodeService()

type inviteCodeService struct {
	failureLimiter inviteCodeFailureLimiter
}

func newInviteCodeService() *inviteCodeService {
	return &inviteCodeService{}
}

func (s *inviteCodeService) Get(id int64) *models.InviteCode {
	return repositories.InviteCodeRepository.Get(sqls.DB(), id)
}

func (s *inviteCodeService) FindPageByCnd(cnd *sqls.Cnd) ([]models.InviteCode, *sqls.Paging) {
	return repositories.InviteCodeRepository.FindPageByCnd(sqls.DB(), cnd)
}

// Generate creates cryptographically random, zero-padded six-digit codes.
func (s *inviteCodeService) Generate(createdByUserId int64, count int, note string) ([]models.InviteCode, error) {
	if createdByUserId <= 0 {
		return nil, errors.New("creator is required")
	}
	if count <= 0 {
		count = 1
	}
	if count > maxInviteCodeBatchSize {
		return nil, fmt.Errorf("batch size must not exceed %d", maxInviteCodeBatchSize)
	}

	config := SysConfigService.GetInviteCodeConfig()
	now := dates.NowTimestamp()
	expiresAt := now + int64(config.ExpireDays)*24*60*60*1000
	note = strings.TrimSpace(note)
	codes := make([]models.InviteCode, 0, count)
	generated := make(map[string]struct{}, count)

	for len(codes) < count {
		code, err := randomInviteCode()
		if err != nil {
			return nil, err
		}
		if _, exists := generated[code]; exists {
			continue
		}
		if repositories.InviteCodeRepository.GetByCode(sqls.DB(), code) != nil {
			continue
		}
		generated[code] = struct{}{}
		codes = append(codes, models.InviteCode{
			Code:            code,
			Status:          constants.InviteCodeStatusUnused,
			CreatedByUserId: createdByUserId,
			Note:            note,
			ExpiresAt:       expiresAt,
			CreateTime:      now,
			UpdateTime:      now,
		})
	}

	err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		for i := range codes {
			if err := repositories.InviteCodeRepository.Create(tx, &codes[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// Consume atomically validates and occupies one unused invite code.
func (s *inviteCodeService) Consume(tx *gorm.DB, code string, usedByUserId int64) (*models.InviteCode, error) {
	code = strings.TrimSpace(code)
	if !validInviteCodeFormat(code) || usedByUserId <= 0 {
		return nil, ErrInviteCodeInvalid
	}

	// Fetch by normalized shape first, then compare the stored value in constant
	// time so validation does not branch on matching bytes.
	stored := repositories.InviteCodeRepository.GetByCode(tx, code)
	if stored == nil ||
		subtle.ConstantTimeCompare([]byte(stored.Code), []byte(code)) != 1 ||
		stored.Status != constants.InviteCodeStatusUnused ||
		stored.UsedByUserId != 0 {
		return nil, ErrInviteCodeInvalid
	}

	now := dates.NowTimestamp()
	if stored.ExpiresAt <= now {
		_ = repositories.InviteCodeRepository.UpdateColumn(tx, stored.Id, "status", constants.InviteCodeStatusExpired)
		return nil, ErrInviteCodeInvalid
	}

	result := tx.Model(&models.InviteCode{}).
		Where("id = ? AND status = ? AND used_by_user_id = 0 AND expires_at > ?", stored.Id, constants.InviteCodeStatusUnused, now).
		Updates(map[string]interface{}{
			"status":          constants.InviteCodeStatusUsed,
			"used_by_user_id": usedByUserId,
			"used_at":         now,
			"update_time":     now,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, ErrInviteCodeInvalid
	}

	stored.Status = constants.InviteCodeStatusUsed
	stored.UsedByUserId = usedByUserId
	stored.UsedAt = now
	stored.UpdateTime = now
	return stored, nil
}

// Disable only changes codes that have not been consumed.
func (s *inviteCodeService) Disable(id, operatorUserId int64, r *http.Request) error {
	code := repositories.InviteCodeRepository.Get(sqls.DB(), id)
	if code == nil {
		return errors.New(locales.Get("admin.entity_not_found"))
	}
	if code.Status != constants.InviteCodeStatusUnused || code.UsedByUserId != 0 {
		return errors.New(locales.Get("user.invite_code_cannot_disable"))
	}
	now := dates.NowTimestamp()
	if err := repositories.InviteCodeRepository.UpdateColumn(sqls.DB(), id, "status", constants.InviteCodeStatusDisabled); err != nil {
		return err
	}
	_ = repositories.InviteCodeRepository.UpdateColumn(sqls.DB(), id, "update_time", now)
	OperateLogService.AddOperateLog(operatorUserId, constants.OpTypeDisable, constants.EntityInviteCode, id, "", r)
	return nil
}

func (s *inviteCodeService) CheckFailureLimit(ip, email, username string) error {
	config := SysConfigService.GetInviteCodeConfig()
	return s.failureLimiter.check(config, inviteCodeFailureKeys(ip, email, username)...)
}

func (s *inviteCodeService) RecordFailure(ip, email, username string) {
	config := SysConfigService.GetInviteCodeConfig()
	s.failureLimiter.record(config, inviteCodeFailureKeys(ip, email, username)...)
}

func (s *inviteCodeService) ResetFailures(ip, email, username string) {
	s.failureLimiter.reset(inviteCodeFailureKeys(ip, email, username)...)
}

func randomInviteCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func validInviteCodeFormat(code string) bool {
	if len(code) != inviteCodeLength {
		return false
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func inviteCodeFailureKeys(ip, email, username string) []string {
	keys := make([]string, 0, 3)
	if value := strings.TrimSpace(ip); value != "" {
		keys = append(keys, "ip:"+hashInviteFailureKey(value))
	}
	if value := strings.TrimSpace(strings.ToLower(email)); value != "" {
		keys = append(keys, "email:"+hashInviteFailureKey(value))
	}
	if value := strings.TrimSpace(strings.ToLower(username)); value != "" {
		keys = append(keys, "username:"+hashInviteFailureKey(value))
	}
	return keys
}

func hashInviteFailureKey(value string) string {
	sum := sha256Sum(value)
	return hex.EncodeToString(sum[:])
}

type inviteCodeFailureState struct {
	failures    []time.Time
	lockedUntil time.Time
}

type inviteCodeFailureLimiter struct {
	mu     sync.Mutex
	states map[string]inviteCodeFailureState
}

func (l *inviteCodeFailureLimiter) check(config dto.InviteCodeConfig, keys ...string) error {
	if config.FailureLimit <= 0 {
		return nil
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ensure()

	for _, key := range keys {
		state := l.states[key]
		if now.Before(state.lockedUntil) {
			return ErrInviteCodeRateLimited
		}
	}
	return nil
}

func (l *inviteCodeFailureLimiter) record(config dto.InviteCodeConfig, keys ...string) {
	if config.FailureLimit <= 0 {
		return
	}
	now := time.Now()
	windowStart := now.Add(-time.Duration(config.FailureWindowMinutes) * time.Minute)
	lockUntil := now.Add(time.Duration(config.LockoutMinutes) * time.Minute)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.ensure()
	for _, key := range keys {
		state := l.states[key]
		failures := state.failures[:0]
		for _, failure := range state.failures {
			if failure.After(windowStart) {
				failures = append(failures, failure)
			}
		}
		failures = append(failures, now)
		if len(failures) >= config.FailureLimit {
			state.lockedUntil = lockUntil
		}
		state.failures = failures
		l.states[key] = state
	}
}

func (l *inviteCodeFailureLimiter) reset(keys ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ensure()
	for _, key := range keys {
		delete(l.states, key)
	}
}

func (l *inviteCodeFailureLimiter) ensure() {
	if l.states == nil {
		l.states = make(map[string]inviteCodeFailureState)
	}
	// Bound memory by dropping expired entries opportunistically.
	if len(l.states) > 10_000 {
		now := time.Now()
		for key, state := range l.states {
			keep := now.Before(state.lockedUntil)
			if !keep && len(state.failures) > 0 {
				keep = state.failures[len(state.failures)-1].After(now.Add(-24 * time.Hour))
			}
			if !keep {
				delete(l.states, key)
			}
		}
	}
}

func sha256Sum(value string) [32]byte {
	return sha256.Sum256([]byte(value))
}
