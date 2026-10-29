package admin

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/repositories"
	"bbs-go/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
)

func agentAdminItem(agent *models.Agent) map[string]interface{} {
	builder := web.NewRspBuilder(agent)
	if owner := services.UserService.Get(agent.OwnerUserId); owner != nil {
		builder.Put("owner", owner.Nickname)
	}
	builder.Put("tokenCount", len(services.AgentService.ListTokens(agent.OwnerUserId, agent.Id)))
	return builder.Build()
}

func buildAgentAdminItem(agent *models.Agent) map[string]interface{} {
	return agentAdminItem(agent)
}

func AgentAdminList(ctx *gin.Context) {
	cnd := params.NewSqlCnd(ctx,
		params.QueryFilter{ParamName: "ownerUserId"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "status"},
	)
	list := repositories.AgentRepository.Find(sqls.DB(), cnd.Desc("id"))
	items := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		items = append(items, agentAdminItem(&list[i]))
	}
	ginx.WriteJSON(ctx, items)
}

func AgentAdminDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	agent := services.AgentService.Get(id)
	if agent == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("agent not found"))
		return
	}
	builder := web.NewRspBuilder(agentAdminItem(agent))
	builder.Put("tokens", services.AgentService.ListTokens(agent.OwnerUserId, agent.Id))
	ginx.WriteJSON(ctx, builder.Build())
}

func AgentAdminDisable(ctx *gin.Context) {
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := services.AgentService.UpdateAdminStatus(id, constants.AgentStatusDisabled); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)
}
