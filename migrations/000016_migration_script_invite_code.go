package migrations

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"bbs-go/internal/pkg/config"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/sqls"
)

func migrate_invite_code_config() error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		tx := ctx.Tx
		if repositories.SysConfigRepository.GetByKey(tx, constants.SysConfigInviteCodeConfig) != nil {
			return nil
		}

		value, err := jsons.ToStr(dto.DefaultInviteCodeConfig())
		if err != nil {
			return err
		}
		name, description := inviteCodeConfigMetaByLanguage()
		return repositories.SysConfigRepository.Create(tx, &models.SysConfig{
			Key:         constants.SysConfigInviteCodeConfig,
			Value:       value,
			Name:        name,
			Description: description,
			CreateTime:  dates.NowTimestamp(),
			UpdateTime:  dates.NowTimestamp(),
		})
	})
}

func inviteCodeConfigMetaByLanguage() (name, description string) {
	if config.Instance != nil && config.Instance.Language == config.LanguageEnUS {
		return "Invite Code Config", "Registration invite-code settings (enable, expiry, failed-attempt limits, and lockout)"
	}
	return "邀请码注册配置", "注册邀请码设置（启用开关、有效期、失败次数限制和锁定时间）"
}
