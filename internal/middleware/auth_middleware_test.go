package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/config"
	"bbs-go/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func TestAuthMiddlewareFallsBackFromAgentBearerToUserToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:auth_middleware_test?mode=memory&cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: "t_", SingularTable: true},
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	sqls.SetDB(db)
	if err := db.AutoMigrate(&models.User{}, &models.UserToken{}, &models.Agent{}, &models.AgentToken{}, &models.SysConfig{}); err != nil {
		t.Fatal(err)
	}
	config.Instance = &config.Config{Installed: true, Language: config.LanguageEnUS}

	user := &models.User{Model: models.Model{Id: 10}, Status: constants.StatusOk}
	if err := repositories.UserRepository.Create(db, user); err != nil {
		t.Fatal(err)
	}
	token := &models.UserToken{Token: "user-token", UserId: user.Id, ExpiredAt: 1 << 62, Status: constants.StatusOk}
	if err := repositories.UserTokenRepository.Create(db, token); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/current", nil)
	ctx.Request.Header.Set("Authorization", "Bearer user-token")
	AuthMiddleware(ctx)

	if got := common.GetCurrentUser(ctx); got == nil || got.Id != user.Id {
		t.Fatalf("expected user token fallback, got %#v", got)
	}
}
