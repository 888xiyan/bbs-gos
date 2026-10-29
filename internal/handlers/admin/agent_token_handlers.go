package admin

import (
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/repositories"
	"bbs-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
)

func AgentTokenAdminList(ctx *gin.Context) {
	cnd := params.NewSqlCnd(ctx,
		params.QueryFilter{ParamName: "agentId"},
		params.QueryFilter{ParamName: "status"},
	)
	list := repositories.AgentTokenRepository.Find(sqls.DB(), cnd.Desc("id"))
	items := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		token := list[i]
		builder := web.NewRspBuilder(token)
		if agent := services.AgentService.Get(token.AgentId); agent != nil {
			builder.Put("agentName", agent.Name)
			builder.Put("ownerUserId", agent.OwnerUserId)
		}
		items = append(items, builder.Build())
	}
	ginx.WriteJSON(ctx, items)
}

func AgentTokenAdminRevoke(ctx *gin.Context) {
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	token := repositories.AgentTokenRepository.Get(sqls.DB(), id)
	if token == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("agent token not found"))
		return
	}
	if err := services.AgentService.RevokeAdminToken(id); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)
}
