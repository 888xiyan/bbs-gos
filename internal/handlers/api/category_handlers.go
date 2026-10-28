package api

import (
	"bbs-go/internal/handlers/render"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/strs"
)

func CategoryTree(ctx *gin.Context) {
	categories := services.CategoryService.GetCategories()
	ginx.WriteJSON(ctx, render.BuildCategoryResponseTree(0, categories))
}

func CategoryByUrl(ctx *gin.Context) {
	urlPath := params.QueryValue(ctx, "urlPath")
	if strs.IsBlank(urlPath) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("urlPath is required"))
		return
	}
	category := services.CategoryService.GetByUrlPath(urlPath)
	if category == nil {
		ginx.WriteJSON(ctx, ginx.ErrorCode(404, "category not found"))
		return
	}
	ginx.WriteJSON(ctx, render.BuildCategory(category))
}
