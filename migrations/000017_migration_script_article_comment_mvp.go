package migrations

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"bbs-go/internal/repositories"
	"fmt"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/sqls"
)

func migrate_article_comment_mvp() error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		tx := ctx.Tx
		now := dates.NowTimestamp()

		modules := dto.ModulesConfig{Article: true}
		modulesJSON, err := jsons.ToStr(modules)
		if err != nil {
			return err
		}
		if existing := repositories.SysConfigRepository.GetByKey(tx, constants.SysConfigModules); existing != nil {
			existing.Value = modulesJSON
			existing.UpdateTime = now
			if err := repositories.SysConfigRepository.Update(tx, existing); err != nil {
				return err
			}
		} else if err := repositories.SysConfigRepository.Create(tx, &models.SysConfig{
			Key:        constants.SysConfigModules,
			Value:      modulesJSON,
			Name:       "功能模块",
			CreateTime: now,
			UpdateTime: now,
		}); err != nil {
			return err
		}

		if existing := repositories.SysConfigRepository.GetByKey(tx, constants.SysConfigCommentsEnabled); existing == nil {
			if err := repositories.SysConfigRepository.Create(tx, &models.SysConfig{
				Key:        constants.SysConfigCommentsEnabled,
				Value:      "true",
				Name:       "全站评论",
				CreateTime: now,
				UpdateTime: now,
			}); err != nil {
				return err
			}
		}

		var categories []models.Category
		if err := tx.Order("id asc").Find(&categories).Error; err != nil {
			return err
		}
		for i := range categories {
			if categories[i].UrlPath == "" {
				categories[i].UrlPath = fmt.Sprintf("/category-%d", categories[i].Id)
				if err := tx.Save(&categories[i]).Error; err != nil {
					return err
				}
			}
		}
		if err := tx.Model(&models.Category{}).Where("1 = 1").Update("allow_post", true).Error; err != nil {
			return err
		}

		defaultCategoryId := int64(0)
		if config := repositories.SysConfigRepository.GetByKey(tx, constants.SysConfigDefaultCategoryId); config != nil {
			fmt.Sscanf(config.Value, "%d", &defaultCategoryId)
		}
		if defaultCategoryId <= 0 && len(categories) > 0 {
			defaultCategoryId = categories[0].Id
		}
		if defaultCategoryId > 0 {
			if err := tx.Model(&models.Article{}).
				Where("category_id = 0 OR category_id IS NULL").
				Update("category_id", defaultCategoryId).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.Article{}).
			Where("1 = 1").
			Update("allow_comments", true).Error; err != nil {
			return err
		}
		return nil
	})
}
