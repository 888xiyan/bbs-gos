package migrations

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
)

func migrate_agent_mvp() error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		tx := ctx.Tx
		now := dates.NowTimestamp()

		for _, seed := range []struct {
			key, value, name string
		}{
			{constants.SysConfigAgentPublishEnabled, "true", "Agent 发布"},
			{constants.SysConfigAgentMaxCount, "1", "每用户 Agent 数量"},
			{constants.SysConfigAgentTokenExpireDays, "90", "Agent Token 有效天数"},
			{constants.SysConfigAgentMandatoryTags, "[]", "Agent 帖子强制标签"},
			{constants.SysConfigAgentConfig, `{"enabled":true,"maxCount":1,"tokenExpireDays":90,"mandatoryTags":[]}`, "Agent 配置"},
		} {
			if existing := repositories.SysConfigRepository.GetByKey(tx, seed.key); existing == nil {
				if err := repositories.SysConfigRepository.Create(tx, &models.SysConfig{
					Key: seed.key, Value: seed.value, Name: seed.name, CreateTime: now, UpdateTime: now,
				}); err != nil {
					return err
				}
			}
		}

		if err := tx.Model(&models.Topic{}).Where("1 = 1").Updates(map[string]interface{}{
			"actor_type":  constants.ActorTypeUser,
			"agent_id":    0,
			"created_via": constants.CreatedViaWeb,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Comment{}).Where("1 = 1").Updates(map[string]interface{}{
			"actor_type":  constants.ActorTypeUser,
			"agent_id":    0,
			"created_via": constants.CreatedViaWeb,
		}).Error; err != nil {
			return err
		}
		return nil
	})
}
