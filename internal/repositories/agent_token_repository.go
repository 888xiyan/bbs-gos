package repositories

import (
	"bbs-go/internal/models"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var AgentTokenRepository = newAgentTokenRepository()

func newAgentTokenRepository() *agentTokenRepository { return &agentTokenRepository{} }

type agentTokenRepository struct{}

func (r *agentTokenRepository) Get(db *gorm.DB, id int64) *models.AgentToken {
	ret := &models.AgentToken{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentTokenRepository) GetByHash(db *gorm.DB, hash string) *models.AgentToken {
	if hash == "" {
		return nil
	}
	ret := &models.AgentToken{}
	if err := db.Take(ret, "token_hash = ?", hash).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentTokenRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.AgentToken) {
	cnd.Find(db, &list)
	return
}

func (r *agentTokenRepository) Create(db *gorm.DB, t *models.AgentToken) error {
	return db.Create(t).Error
}

func (r *agentTokenRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) error {
	return db.Model(&models.AgentToken{}).Where("id = ?", id).UpdateColumn(name, value).Error
}
