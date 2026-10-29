package common

import (
	"bbs-go/internal/models"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/idcodec"
	"bbs-go/internal/pkg/params"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

const (
	current_user_key  = "__current_user"
	current_actor_key = "__current_actor"
)

func SetCurrentUser(ctx *gin.Context, user *models.User) {
	ctx.Set(current_user_key, user)
	ctx.Set(current_actor_key, models.NewUserActor(user))
}

func SetCurrentActor(ctx *gin.Context, actor *models.ActorContext) {
	if actor == nil {
		return
	}
	ctx.Set(current_actor_key, actor)
	if actor.User != nil {
		ctx.Set(current_user_key, actor.User)
	}
}

func GetCurrentActor(ctx *gin.Context) *models.ActorContext {
	if v, exists := ctx.Get(current_actor_key); exists {
		if actor, ok := v.(*models.ActorContext); ok {
			return actor
		}
	}
	return models.NewUserActor(GetCurrentUser(ctx))
}

func GetCurrentUserID(ctx *gin.Context) int64 {
	user := GetCurrentUser(ctx)
	if user != nil {
		return user.Id
	}
	return 0
}

func GetCurrentUser(ctx *gin.Context) *models.User {
	if v, exists := ctx.Get(current_user_key); exists {
		if u, ok := v.(*models.User); ok {
			return u
		}
	}
	return nil
}

func CheckLogin(ctx *gin.Context) (*models.User, error) {
	user := GetCurrentUser(ctx)
	if user == nil {
		return nil, errs.NotLogin()
	}
	return user, nil
}

func IsLogin(ctx *gin.Context) bool {
	return GetCurrentUser(ctx) != nil
}

func GetID(ctx *gin.Context, name string) int64 {
	idStr, _ := params.Get(ctx, name)
	return idcodec.Decode(idStr)
}
