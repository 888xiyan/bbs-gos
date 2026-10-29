package middleware

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/config"
	"bbs-go/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(ctx *gin.Context) {
	if config.Instance.Installed {
		if authorization := ctx.GetHeader("Authorization"); strings.HasPrefix(authorization, "Bearer ") {
			if auth, err := services.AgentService.Authenticate(strings.TrimPrefix(authorization, "Bearer ")); err == nil {
				common.SetCurrentActor(ctx, &models.ActorContext{
					User:          auth.Owner,
					Agent:         auth.Agent,
					AgentToken:    auth.Token,
					ActorType:     constants.ActorTypeAgent,
					OwnerUserId:   auth.Owner.Id,
					AgentId:       auth.Agent.Id,
					Authenticated: true,
				})
			}
		} else if user := services.UserTokenService.GetCurrent(ctx); user != nil {
			common.SetCurrentUser(ctx, user)
		}
	}
	ctx.Next()
}
