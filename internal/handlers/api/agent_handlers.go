package api

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/web"
)

func agentOwner(ctx *gin.Context) (*models.User, bool) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("not login"))
		return nil, false
	}
	return user, true
}

func AgentsCreateHandler(ctx *gin.Context) {
	user, ok := agentOwner(ctx)
	if !ok {
		return
	}
	var form req.CreateAgentReq
	if err := ginx.Bind(ctx, &form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	agent, err := services.AgentService.Create(user.Id, form)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, agent)
}

func AgentsListHandler(ctx *gin.Context) {
	user, ok := agentOwner(ctx)
	if !ok {
		return
	}
	ginx.WriteJSON(ctx, services.AgentService.FindByOwner(user.Id))
}

func AgentDetailRoute(ctx *gin.Context) {
	user, ok := agentOwner(ctx)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	agent := services.AgentService.GetOwned(user.Id, id)
	if agent == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("agent not found"))
		return
	}
	builder := web.NewRspBuilder(agent)
	builder.Put("tokens", services.AgentService.ListTokens(user.Id, agent.Id))
	ginx.WriteJSON(ctx, builder.Build())
}

func AgentsUpdateHandler(ctx *gin.Context) {
	user, ok := agentOwner(ctx)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	var form req.UpdateAgentReq
	if err := ginx.Bind(ctx, &form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	agent, err := services.AgentService.Update(user.Id, id, form)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, agent)
}

func AgentsDeleteHandler(ctx *gin.Context) {
	user, ok := agentOwner(ctx)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := services.AgentService.Disable(user.Id, id); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)
}

func AgentTokensIssueHandler(ctx *gin.Context) {
	user, ok := agentOwner(ctx)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	issue, err := services.AgentService.IssueToken(user.Id, id)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	builder := web.NewRspBuilder(issue.Token)
	builder.Put("token", issue.Value)
	ginx.WriteJSON(ctx, builder.Build())
}

func AgentTokensRevokeHandler(ctx *gin.Context) {
	user, ok := agentOwner(ctx)
	if !ok {
		return
	}
	agentId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	tokenId, err := strconv.ParseInt(ctx.Param("tokenId"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := services.AgentService.RevokeToken(user.Id, agentId, tokenId); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)
}

func AgentTokensListHandler(ctx *gin.Context) {
	user, ok := agentOwner(ctx)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, services.AgentService.ListTokens(user.Id, id))
}

func AgentContextHandler(ctx *gin.Context) {
	actor := common.GetCurrentActor(ctx)
	if actor == nil || !actor.Authenticated {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("not login"))
		return
	}
	ginx.WriteJSON(ctx, map[string]interface{}{
		"actorType":   actor.ActorType,
		"ownerUserId": actor.OwnerUserId,
		"agentId":     actor.AgentId,
	})
}
