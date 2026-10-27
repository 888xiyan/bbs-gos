package services

import (
	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"bbs-go/internal/pkg/config"
	"bbs-go/internal/pkg/search"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func setupInviteCodeTestDB(t *testing.T, enabled bool) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:invite_code_test_%d?mode=memory&cache=shared&_fk=1", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	previousConfig := config.Instance
	t.Cleanup(func() { config.Instance = previousConfig })
	config.Instance = &config.Config{
		Language: config.LanguageZhCN,
		Search:   config.SearchConfig{IndexPath: filepath.Join(t.TempDir(), "search.bleve")},
	}
	search.Init()
	sqls.SetDB(db)

	if err := db.AutoMigrate(
		&models.User{},
		&models.InviteCode{},
		&models.SysConfig{},
		&models.OperateLog{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	inviteConfig := dto.DefaultInviteCodeConfig()
	inviteConfig.Enabled = enabled
	value, err := jsons.ToStr(inviteConfig)
	if err != nil {
		t.Fatalf("marshal invite config: %v", err)
	}
	if err := db.Create(&models.SysConfig{
		Key:        constants.SysConfigInviteCodeConfig,
		Value:      value,
		CreateTime: time.Now().UnixMilli(),
		UpdateTime: time.Now().UnixMilli(),
	}).Error; err != nil {
		t.Fatalf("create invite config: %v", err)
	}
	cache.SysConfigCache.Invalidate(constants.SysConfigInviteCodeConfig)
	return db
}

func TestInviteCodeGenerateCreatesUniqueSixDigitCodes(t *testing.T) {
	setupInviteCodeTestDB(t, false)

	codes, err := InviteCodeService.Generate(1001, 25, "test")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(codes) != 25 {
		t.Fatalf("expected 25 codes, got %d", len(codes))
	}
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		if len(code.Code) != 6 || strings.Trim(code.Code, "0123456789") != "" {
			t.Fatalf("expected six digits, got %q", code.Code)
		}
		if _, exists := seen[code.Code]; exists {
			t.Fatalf("duplicate code %q", code.Code)
		}
		seen[code.Code] = struct{}{}
		if code.Status != constants.InviteCodeStatusUnused || code.UsedByUserId != 0 {
			t.Fatalf("unexpected generated state: %+v", code)
		}
	}
}

func TestSignUpWithInviteCodeConsumesCodeAndRollsBackOnInvalidCode(t *testing.T) {
	db := setupInviteCodeTestDB(t, true)
	codes, err := InviteCodeService.Generate(1001, 1, "signup")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	user, err := UserService.SignUpWithInviteCode(
		"invited-user", "invited@example.com", "Invited", "password", "password",
		codes[0].Code, "127.0.0.1", true,
	)
	if err != nil {
		t.Fatalf("sign up: %v", err)
	}
	if user.InviteCodeId != codes[0].Id || user.InvitedByUserId != 1001 {
		t.Fatalf("expected invite linkage, got user=%+v", user)
	}

	var used models.InviteCode
	if err := db.First(&used, "id = ?", codes[0].Id).Error; err != nil {
		t.Fatalf("load used invite: %v", err)
	}
	if used.Status != constants.InviteCodeStatusUsed || used.UsedByUserId != user.Id {
		t.Fatalf("expected consumed invite, got %+v", used)
	}

	if _, err := UserService.SignUpWithInviteCode(
		"rollback-user", "rollback@example.com", "Rollback", "password", "password",
		"000000", "127.0.0.2", true,
	); err == nil {
		t.Fatal("expected invalid invite signup to fail")
	}

	var rollbackCount int64
	if err := db.Model(&models.User{}).Where("email = ?", "rollback@example.com").Count(&rollbackCount).Error; err != nil {
		t.Fatalf("count rollback user: %v", err)
	}
	if rollbackCount != 0 {
		t.Fatalf("expected user creation to roll back, got %d rows", rollbackCount)
	}
}

func TestInviteCodeConsumeRejectsExpiredDisabledAndUsedCodes(t *testing.T) {
	db := setupInviteCodeTestDB(t, true)
	now := time.Now().UnixMilli()
	for _, seed := range []models.InviteCode{
		{Code: "111111", Status: constants.InviteCodeStatusUnused, CreatedByUserId: 1, ExpiresAt: now - 1, CreateTime: now, UpdateTime: now},
		{Code: "222222", Status: constants.InviteCodeStatusDisabled, CreatedByUserId: 1, ExpiresAt: now + 60_000, CreateTime: now, UpdateTime: now},
		{Code: "333333", Status: constants.InviteCodeStatusUsed, CreatedByUserId: 1, UsedByUserId: 2, ExpiresAt: now + 60_000, UsedAt: now, CreateTime: now, UpdateTime: now},
	} {
		seed := seed
		if err := db.Create(&seed).Error; err != nil {
			t.Fatalf("create seed invite: %v", err)
		}
		if _, err := InviteCodeService.Consume(db, seed.Code, 99); err == nil {
			t.Fatalf("expected code %q to be rejected", seed.Code)
		}
	}

	var expired models.InviteCode
	if err := db.First(&expired, "code = ?", "111111").Error; err != nil {
		t.Fatalf("load expired invite: %v", err)
	}
	if expired.Status != constants.InviteCodeStatusExpired {
		t.Fatalf("expected expired status, got %q", expired.Status)
	}
}

func TestInviteCodeFailureLimiterLocksRepeatedFailures(t *testing.T) {
	setupInviteCodeTestDB(t, true)
	limiter := newInviteCodeService()
	config := dto.DefaultInviteCodeConfig()
	config.Enabled = true
	config.FailureLimit = 3
	config.FailureWindowMinutes = 10
	config.LockoutMinutes = 5

	for i := 0; i < config.FailureLimit; i++ {
		limiter.failureLimiter.record(config, inviteCodeFailureKeys("10.0.0.1", "brute@example.com", "brute")...)
	}
	if err := limiter.failureLimiter.check(config, inviteCodeFailureKeys("10.0.0.1", "brute@example.com", "brute")...); err == nil {
		t.Fatal("expected limiter to lock after repeated failures")
	}
	if err := limiter.failureLimiter.check(config, inviteCodeFailureKeys("10.0.0.2", "other@example.com", "other")...); err != nil {
		t.Fatalf("expected unrelated identity to remain allowed: %v", err)
	}
}
