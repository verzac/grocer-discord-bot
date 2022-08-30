package repositories

import (
	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ GrohereRecordRepository = &GrohereRecordRepositoryImpl{}

type GrohereRecordQueryOpts struct {
	IsStrongNilForGroceryListID bool
}

type GrohereRecordRepository interface {
	FindByQueryWithConfig(q *models.GrohereRecord, config GrohereRecordQueryOpts) ([]models.GrohereRecord, error)
	FindByQuery(q *models.GrohereRecord) ([]models.GrohereRecord, error)
	Put(g *models.GrohereRecord) error
	Delete(g *models.GrohereRecord) error
}

type GrohereRecordRepositoryImpl struct {
	DB *gorm.DB
}

func (r *GrohereRecordRepositoryImpl) FindByQueryWithConfig(q *models.GrohereRecord, config GrohereRecordQueryOpts) ([]models.GrohereRecord, error) {
	entries := make([]models.GrohereRecord, 0)
	dbQuery := r.DB.Where(q)
	if config.IsStrongNilForGroceryListID && q.GroceryListID == nil {
		dbQuery = dbQuery.Where(queryGroceryListIDIsNil)
	}
	if res := dbQuery.Find(&entries); res.Error != nil {
		if res.Error != gorm.ErrRecordNotFound {
			return nil, res.Error
		}
	}
	return entries, nil
}

func (r *GrohereRecordRepositoryImpl) FindByQuery(q *models.GrohereRecord) ([]models.GrohereRecord, error) {
	return r.FindByQueryWithConfig(q, GrohereRecordQueryOpts{})
}

func (r *GrohereRecordRepositoryImpl) Put(g *models.GrohereRecord) error {
	if r := r.DB.Save(g); r.Error != nil {
		return r.Error
	}
	return nil
}

func (r *GrohereRecordRepositoryImpl) Delete(g *models.GrohereRecord) error {
	return r.DB.Delete(g).Error
}
