package repositories

import (
	"bbs-go/internal/models"
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var InviteCodeRepository = newInviteCodeRepository()

func newInviteCodeRepository() *inviteCodeRepository {
	return &inviteCodeRepository{}
}

type inviteCodeRepository struct{}

func (r *inviteCodeRepository) Get(db *gorm.DB, id int64) *models.InviteCode {
	ret := &models.InviteCode{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *inviteCodeRepository) GetByCode(db *gorm.DB, code string) *models.InviteCode {
	if len(code) == 0 {
		return nil
	}
	ret := &models.InviteCode{}
	if err := db.Take(ret, "code = ?", code).Error; err != nil {
		return nil
	}
	return ret
}

func (r *inviteCodeRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.InviteCode) {
	cnd.Find(db, &list)
	return
}

func (r *inviteCodeRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.InviteCode {
	ret := &models.InviteCode{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *inviteCodeRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.InviteCode, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *inviteCodeRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.InviteCode, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.InviteCode{})
	paging = &sqls.Paging{Page: cnd.Paging.Page, Limit: cnd.Paging.Limit, Total: count}
	return
}

func (r *inviteCodeRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.InviteCode{})
}

func (r *inviteCodeRepository) Create(db *gorm.DB, t *models.InviteCode) error {
	return db.Create(t).Error
}

func (r *inviteCodeRepository) Update(db *gorm.DB, t *models.InviteCode) error {
	return db.Save(t).Error
}

func (r *inviteCodeRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) error {
	return db.Model(&models.InviteCode{}).Where("id = ?", id).Updates(columns).Error
}

func (r *inviteCodeRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) error {
	return db.Model(&models.InviteCode{}).Where("id = ?", id).UpdateColumn(name, value).Error
}
