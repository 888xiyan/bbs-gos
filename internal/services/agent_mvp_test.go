package services

import (
	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/repositories"
	"testing"
	"time"

	"github.com/mlogclub/simple/sqls"
)

func setupAgentMVPTestDB(t *testing.T) {
	t.Helper()
	db := setupTestDB(t)
	if err := db.AutoMigrate(
		&models.User{},
		&models.Agent{},
		&models.AgentToken{},
		&models.Topic{},
		&models.TopicTag{},
		&models.Tag{},
		&models.Category{},
		&models.SysConfig{},
		&models.OperateLog{},
		&models.ForbiddenWord{},
	); err != nil {
		t.Fatalf("auto migrate agent mvp: %v", err)
	}
}

func TestTopicPublishServiceMergesAgentMandatoryTags(t *testing.T) {
	setupAgentMVPTestDB(t)
	tag := &models.Tag{Model: models.Model{Id: 100}, Name: "Required"}
	if err := repositories.TagRepository.Create(sqls.DB(), tag); err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if err := sqls.DB().Create(&models.SysConfig{
		Key:   constants.SysConfigAgentMandatoryTags,
		Value: "[100]",
	}).Error; err != nil {
		t.Fatalf("create config: %v", err)
	}
	cache.SysConfigCache.Invalidate(constants.SysConfigAgentMandatoryTags)

	tags := TopicPublishService.mergeAgentMandatoryTags([]string{"requested"})
	if len(tags) != 2 || tags[0] != "requested" || tags[1] != "Required" {
		t.Fatalf("unexpected merged tags: %#v", tags)
	}
}

func TestAgentTokenAuthenticationAndMandatoryTags(t *testing.T) {
	setupAgentMVPTestDB(t)
	owner := mustCreateUser(t, time.Now().UnixMilli())
	agent, err := AgentService.Create(owner.Id, req.CreateAgentReq{Name: "bot"})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	issue, err := AgentService.IssueToken(owner.Id, agent.Id)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	var rawStored int64
	if err := sqls.DB().Model(&models.AgentToken{}).
		Where("token_hash = ? OR token_prefix = ?", issue.Value, issue.Value).
		Count(&rawStored).Error; err != nil {
		t.Fatalf("inspect raw token storage: %v", err)
	}
	if rawStored != 0 {
		t.Fatal("raw agent token must not be stored")
	}
	auth, err := AgentService.Authenticate(issue.Value)
	if err != nil {
		t.Fatalf("authenticate token: %v", err)
	}
	if auth.Agent.Id != agent.Id || auth.Owner.Id != owner.Id {
		t.Fatalf("unexpected auth context: %+v", auth)
	}
	if err := AgentService.RevokeToken(owner.Id, agent.Id, issue.Token.Id); err != nil {
		t.Fatalf("revoke token: %v", err)
	}
	if _, err := AgentService.Authenticate(issue.Value); err == nil {
		t.Fatal("expected revoked token authentication to fail")
	}

}
