package admin

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/web"
)

type inviteCodeCreateReq struct {
	Count int    `json:"count" form:"count"`
	Note  string `json:"note" form:"note"`
}

func inviteCodeStatus(code *models.InviteCode) string {
	if code.Status == constants.InviteCodeStatusUnused && code.ExpiresAt <= dates.NowTimestamp() {
		return constants.InviteCodeStatusExpired
	}
	return code.Status
}

func inviteCodeBuildItem(code *models.InviteCode) map[string]interface{} {
	builder := web.NewRspBuilder(code)
	builder.Put("status", inviteCodeStatus(code))
	if creator := services.UserService.Get(code.CreatedByUserId); creator != nil {
		builder.Put("createdBy", creator.Nickname)
	} else {
		builder.Put("createdBy", "")
	}
	if code.UsedByUserId > 0 {
		if usedBy := services.UserService.Get(code.UsedByUserId); usedBy != nil {
			builder.Put("usedBy", usedBy.Nickname)
		} else {
			builder.Put("usedBy", "")
		}
	} else {
		builder.Put("usedBy", "")
	}
	return builder.Build()
}

func InviteCodeList(ctx *gin.Context) {
	cnd := params.NewPagedSqlCnd(ctx,
		params.QueryFilter{ParamName: "code", Op: params.Like},
		params.QueryFilter{ParamName: "createdByUserId"},
		params.QueryFilter{ParamName: "usedByUserId"},
	)
	if start := params.GetTime(ctx, "usedStartTime"); start != nil {
		cnd.Gte("used_at", start.UnixMilli())
	}
	if end := params.GetTime(ctx, "usedEndTime"); end != nil {
		cnd.Lte("used_at", end.UnixMilli()+24*60*60*1000-1)
	}
	now := dates.NowTimestamp()
	switch params.QueryValue(ctx, "status") {
	case constants.InviteCodeStatusUnused:
		cnd.Eq("status", constants.InviteCodeStatusUnused).Gt("expires_at", now)
	case constants.InviteCodeStatusExpired:
		cnd.Eq("status", constants.InviteCodeStatusUnused).Lte("expires_at", now)
	case constants.InviteCodeStatusUsed, constants.InviteCodeStatusDisabled:
		cnd.Eq("status", params.QueryValue(ctx, "status"))
	}

	list, paging := services.InviteCodeService.FindPageByCnd(cnd.Desc("id"))
	items := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		items = append(items, inviteCodeBuildItem(&list[i]))
	}

	if user := common.GetCurrentUser(ctx); user != nil {
		services.OperateLogService.AddOperateLog(
			user.Id,
			constants.OpTypeView,
			constants.EntityInviteCode,
			0,
			"view invite code list",
			ctx.Request,
		)
	}
	ginx.WriteJSON(ctx, &web.PageResult{Results: items, Page: paging})
}

func InviteCodeCreate(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("not login"))
		return
	}
	var req inviteCodeCreateReq
	if err := ginx.Bind(ctx, &req); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if req.Count == 0 {
		req.Count = 10
	}
	codes, err := services.InviteCodeService.Generate(user.Id, req.Count, req.Note)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(
		user.Id,
		constants.OpTypeGenerate,
		constants.EntityInviteCode,
		0,
		"generated "+strconv.Itoa(len(codes))+" invite codes",
		ctx.Request,
	)
	items := make([]map[string]interface{}, 0, len(codes))
	for i := range codes {
		items = append(items, inviteCodeBuildItem(&codes[i]))
	}
	ginx.WriteJSON(ctx, items)
}

func InviteCodeDisable(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("not login"))
		return
	}
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := services.InviteCodeService.Disable(id, user.Id, ctx.Request); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)
}

func InviteCodeDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	code := services.InviteCodeService.Get(id)
	if code == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("entity not found"))
		return
	}
	ginx.WriteJSON(ctx, inviteCodeBuildItem(code))
}
