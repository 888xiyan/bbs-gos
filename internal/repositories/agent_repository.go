package repositories

import (
	"bbs-go/internal/models"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var AgentRepository = newAgentRepository()

func newAgentRepository() *agentRepository { return &agentRepository{} }

type agentRepository struct{}

func (r *agentRepository) Get(db *gorm.DB, id int64) *models.Agent {
	ret := &models.Agent{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Agent) {
	cnd.Find(db, &list)
	return
}

func (r *agentRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.Agent {
	ret := &models.Agent{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *agentRepository) Create(db *gorm.DB, t *models.Agent) error {
	return db.Create(t).Error
}

func (r *agentRepository) Update(db *gorm.DB, t *models.Agent) error {
	return db.Save(t).Error
}

func (r *agentRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) error {
	return db.Model(&models.Agent{}).Where("id = ?", id).Updates(columns).Error
}

func (r *agentRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) error {
	return db.Model(&models.Agent{}).Where("id = ?", id).UpdateColumn(name, value).Error
}
