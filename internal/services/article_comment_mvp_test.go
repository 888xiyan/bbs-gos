package services

import (
	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/repositories"
	"strconv"
	"testing"
	"time"

	"github.com/mlogclub/simple/sqls"
)

func setupArticleCommentMVPTestDB(t *testing.T) {
	t.Helper()
	db := setupTestDB(t)
	if err := db.AutoMigrate(
		&models.Article{},
		&models.Comment{},
		&models.User{},
		&models.Category{},
		&models.SysConfig{},
	); err != nil {
		t.Fatalf("auto migrate article comment mvp: %v", err)
	}
	for _, key := range []string{constants.SysConfigCommentsEnabled, constants.SysConfigAntiAbuseConfig} {
		if err := db.Create(&models.SysConfig{Key: key, Value: ""}).Error; err != nil {
			t.Fatalf("create sys config %s: %v", key, err)
		}
		cache.SysConfigCache.Invalidate(key)
	}
}

func TestCommentServicePublishesArticleCommentsAndRepliesWithCorrectCounts(t *testing.T) {
	setupArticleCommentMVPTestDB(t)
	now := time.Now().UnixMilli()
	article := &models.Article{
		UserId:        10,
		CategoryId:    20,
		Title:         "article",
		Content:       "content",
		ContentType:   constants.ContentTypeText,
		Status:        constants.StatusOk,
		AllowComments: true,
		CreateTime:    now,
		UpdateTime:    now,
	}
	if err := repositories.ArticleRepository.Create(sqls.DB(), article); err != nil {
		t.Fatalf("create article: %v", err)
	}

	topLevel, err := CommentService.Publish(11, req.CreateCommentReq{
		EntityType: constants.EntityArticle,
		EntityId:   strconv.FormatInt(article.Id, 10),
		Content:    "top-level",
	})
	if err != nil {
		t.Fatalf("publish top-level comment: %v", err)
	}

	reply, err := CommentService.Publish(12, req.CreateCommentReq{
		EntityType: constants.EntityComment,
		EntityId:   strconv.FormatInt(topLevel.Id, 10),
		Content:    "reply",
	})
	if err != nil {
		t.Fatalf("publish reply: %v", err)
	}

	gotArticle := ArticleService.Get(article.Id)
	if gotArticle.CommentCount != 2 {
		t.Fatalf("expected article comment count 2, got %d", gotArticle.CommentCount)
	}
	gotParent := CommentService.Get(topLevel.Id)
	if gotParent.CommentCount != 1 {
		t.Fatalf("expected parent comment count 1, got %d", gotParent.CommentCount)
	}
	if reply.EntityType != constants.EntityComment {
		t.Fatalf("expected reply entity comment, got %q", reply.EntityType)
	}
}

func TestCommentServiceRejectsDisabledArticleAndNonArticleEntities(t *testing.T) {
	setupArticleCommentMVPTestDB(t)
	now := time.Now().UnixMilli()
	article := &models.Article{
		UserId:        10,
		CategoryId:    20,
		Title:         "article",
		Content:       "content",
		ContentType:   constants.ContentTypeText,
		Status:        constants.StatusOk,
		AllowComments: false,
		CreateTime:    now,
		UpdateTime:    now,
	}
	if err := repositories.ArticleRepository.Create(sqls.DB(), article); err != nil {
		t.Fatalf("create article: %v", err)
	}

	if _, err := CommentService.Publish(11, req.CreateCommentReq{
		EntityType: constants.EntityArticle,
		EntityId:   strconv.FormatInt(article.Id, 10),
		Content:    "blocked",
	}); err == nil {
		t.Fatal("expected disabled article comments to be rejected")
	}
	if _, err := CommentService.Publish(11, req.CreateCommentReq{
		EntityType: constants.EntityTopic,
		EntityId:   strconv.FormatInt(article.Id, 10),
		Content:    "invalid entity",
	}); err == nil {
		t.Fatal("expected non-article entity to be rejected")
	}

	if err := sqls.DB().Model(&models.SysConfig{}).
		Where("key = ?", constants.SysConfigCommentsEnabled).
		Update("value", "false").Error; err != nil {
		t.Fatalf("disable comments: %v", err)
	}
	if _, err := CommentService.Publish(11, req.CreateCommentReq{
		EntityType: constants.EntityArticle,
		EntityId:   strconv.FormatInt(article.Id, 10),
		Content:    "globally blocked",
	}); err == nil {
		t.Fatal("expected globally disabled comments to be rejected")
	}
}
