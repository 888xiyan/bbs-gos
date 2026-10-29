package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/repositories"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
)

const (
	agentStatusActive   = "active"
	agentStatusDisabled = "disabled"
	agentStatusRevoked  = "revoked"

	agentTokenStatusActive  = "active"
	agentTokenStatusRevoked = "revoked"

	defaultAgentTokenTTL = 90 * 24 * time.Hour
)

var AgentService = newAgentService()

func newAgentService() *agentService {
	return &agentService{}
}

type agentService struct{}

type AgentAuth struct {
	Owner *models.User
	Agent *models.Agent
	Token *models.AgentToken
}

type AgentTokenIssue struct {
	Token *models.AgentToken
	Value string
}

func (s *agentService) Get(id int64) *models.Agent {
	return repositories.AgentRepository.Get(sqls.DB(), id)
}

func (s *agentService) GetOwned(ownerUserId, id int64) *models.Agent {
	if ownerUserId <= 0 || id <= 0 {
		return nil
	}
	return repositories.AgentRepository.FindOne(sqls.DB(), sqls.NewCnd().Eq("id", id).Eq("owner_user_id", ownerUserId))
}

func (s *agentService) FindByOwner(ownerUserId int64) []models.Agent {
	if ownerUserId <= 0 {
		return nil
	}
	return repositories.AgentRepository.Find(sqls.DB(), sqls.NewCnd().Eq("owner_user_id", ownerUserId).Desc("id"))
}

func (s *agentService) CountByOwner(ownerUserId int64) int {
	if ownerUserId <= 0 {
		return 0
	}
	return len(repositories.AgentRepository.Find(sqls.DB(), sqls.NewCnd().Eq("owner_user_id", ownerUserId).Eq("status", agentStatusActive)))
}

func (s *agentService) Create(ownerUserId int64, form req.CreateAgentReq) (*models.Agent, error) {
	if ownerUserId <= 0 {
		return nil, errors.New("not login")
	}
	if !SysConfigService.GetAgentConfig().Enabled {
		return nil, errors.New("agent publishing is disabled")
	}
	name := strings.TrimSpace(form.Name)
	description := strings.TrimSpace(form.Description)
	if name == "" || len([]rune(name)) > 64 {
		return nil, errors.New("invalid agent name")
	}
	if len([]rune(description)) > 512 {
		return nil, errors.New("invalid agent description")
	}
	if s.CountByOwner(ownerUserId) >= SysConfigService.GetAgentConfig().MaxCount {
		return nil, errors.New("agent limit reached")
	}

	now := dates.NowTimestamp()
	agent := &models.Agent{
		OwnerUserId: ownerUserId,
		Name:        name,
		Description: description,
		Status:      agentStatusActive,
		CreateTime:  now,
		UpdateTime:  now,
	}
	if err := repositories.AgentRepository.Create(sqls.DB(), agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *agentService) Update(ownerUserId, id int64, form req.UpdateAgentReq) (*models.Agent, error) {
	agent := s.GetOwned(ownerUserId, id)
	if agent == nil {
		return nil, errors.New("agent not found")
	}
	name := strings.TrimSpace(form.Name)
	description := strings.TrimSpace(form.Description)
	if name == "" || len([]rune(name)) > 64 {
		return nil, errors.New("invalid agent name")
	}
	if len([]rune(description)) > 512 {
		return nil, errors.New("invalid agent description")
	}
	status := agent.Status
	if form.Status != "" {
		status = strings.ToLower(strings.TrimSpace(form.Status))
		if status != agentStatusActive && status != agentStatusDisabled && status != agentStatusRevoked {
			return nil, errors.New("invalid agent status")
		}
	}
	now := dates.NowTimestamp()
	if err := repositories.AgentRepository.Update(sqls.DB(), &models.Agent{
		Model:          models.Model{Id: agent.Id},
		OwnerUserId:    agent.OwnerUserId,
		Name:           name,
		Description:    description,
		Status:         status,
		LastActiveTime: agent.LastActiveTime,
		CreateTime:     agent.CreateTime,
		UpdateTime:     now,
	}); err != nil {
		return nil, err
	}
	return s.Get(agent.Id), nil
}

func (s *agentService) Disable(ownerUserId, id int64) error {
	agent := s.GetOwned(ownerUserId, id)
	if agent == nil {
		return errors.New("agent not found")
	}
	return repositories.AgentRepository.UpdateColumn(sqls.DB(), id, "status", agentStatusDisabled)
}

func (s *agentService) IssueToken(ownerUserId, agentId int64) (*AgentTokenIssue, error) {
	agent := s.GetOwned(ownerUserId, agentId)
	if agent == nil {
		return nil, errors.New("agent not found")
	}
	if agent.Status != agentStatusActive {
		return nil, errors.New("agent is not active")
	}

	raw := randomAgentToken()
	hash := hashAgentToken(raw)
	now := dates.NowTimestamp()
	token := &models.AgentToken{
		AgentId:     agent.Id,
		TokenHash:   hash,
		TokenPrefix: raw[:8],
		Status:      agentTokenStatusActive,
		ExpiredAt:   dates.Timestamp(time.Now().Add(time.Duration(SysConfigService.GetAgentConfig().TokenExpireDays) * 24 * time.Hour)),
		CreateTime:  now,
	}
	if err := repositories.AgentTokenRepository.Create(sqls.DB(), token); err != nil {
		return nil, err
	}
	OperateLogService.AddOperateLog(ownerUserId, constants.OpTypeCreate, "agentToken", token.Id, "issued agent token", nil)
	return &AgentTokenIssue{Token: token, Value: raw}, nil
}

func (s *agentService) RevokeToken(ownerUserId, agentId, tokenId int64) error {
	agent := s.GetOwned(ownerUserId, agentId)
	if agent == nil {
		return errors.New("agent not found")
	}
	token := repositories.AgentTokenRepository.Get(sqls.DB(), tokenId)
	if token == nil || token.AgentId != agent.Id {
		return errors.New("agent token not found")
	}
	return s.revokeToken(token, ownerUserId)
}

func (s *agentService) revokeToken(token *models.AgentToken, operatorUserId int64) error {
	now := dates.NowTimestamp()
	if err := repositories.AgentTokenRepository.UpdateColumn(sqls.DB(), token.Id, "status", agentTokenStatusRevoked); err != nil {
		return err
	}
	if err := repositories.AgentTokenRepository.UpdateColumn(sqls.DB(), token.Id, "revoked_time", now); err != nil {
		return err
	}
	OperateLogService.AddOperateLog(operatorUserId, constants.OpTypeUpdate, "agentToken", token.Id, "revoked agent token", nil)
	return nil
}

func (s *agentService) Authenticate(tokenValue string) (*AgentAuth, error) {
	tokenValue = strings.TrimSpace(tokenValue)
	if tokenValue == "" {
		return nil, errors.New("agent token required")
	}
	hash := hashAgentToken(tokenValue)
	token := repositories.AgentTokenRepository.GetByHash(sqls.DB(), hash)
	if token == nil || token.Status != agentTokenStatusActive {
		return nil, errors.New("invalid agent token")
	}
	if token.ExpiredAt > 0 && token.ExpiredAt <= dates.NowTimestamp() {
		return nil, errors.New("agent token expired")
	}
	agent := repositories.AgentRepository.Get(sqls.DB(), token.AgentId)
	if agent == nil || agent.Status != agentStatusActive {
		return nil, errors.New("agent is not active")
	}
	owner := repositories.UserRepository.Get(sqls.DB(), agent.OwnerUserId)
	if owner == nil || owner.Status != constants.StatusOk {
		return nil, errors.New("agent owner is unavailable")
	}
	if err := UserService.CheckPostStatus(owner); err != nil {
		return nil, err
	}
	_ = repositories.AgentTokenRepository.UpdateColumn(sqls.DB(), token.Id, "last_used_time", dates.NowTimestamp())
	_ = repositories.AgentRepository.UpdateColumn(sqls.DB(), agent.Id, "last_active_time", dates.NowTimestamp())
	return &AgentAuth{Owner: owner, Agent: agent, Token: token}, nil
}

func (s *agentService) FindByOwnerAndName(ownerUserId int64, name string) *models.Agent {
	if ownerUserId <= 0 {
		return nil
	}
	return repositories.AgentRepository.FindOne(sqls.DB(), sqls.NewCnd().Eq("owner_user_id", ownerUserId).Eq("name", name))
}

func (s *agentService) ListTokens(ownerUserId, agentId int64) []models.AgentToken {
	if s.GetOwned(ownerUserId, agentId) == nil {
		return nil
	}
	return repositories.AgentTokenRepository.Find(sqls.DB(), sqls.NewCnd().Eq("agent_id", agentId).Desc("id"))
}

func randomAgentToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return "bbsa_" + hex.EncodeToString(buf)
}

func hashAgentToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *agentService) UpdateAdminStatus(id int64, status string) error {
	if status != agentStatusActive && status != agentStatusDisabled && status != agentStatusRevoked {
		return errors.New("invalid agent status")
	}
	if err := repositories.AgentRepository.UpdateColumn(sqls.DB(), id, "status", status); err != nil {
		return err
	}
	OperateLogService.AddOperateLog(0, constants.OpTypeUpdate, constants.EntityAgent, id, "updated agent status", nil)
	return nil
}

func (s *agentService) RevokeAdminToken(id int64) error {
	token := repositories.AgentTokenRepository.Get(sqls.DB(), id)
	if token == nil {
		return errors.New("agent token not found")
	}
	return s.revokeToken(token, 0)
}
