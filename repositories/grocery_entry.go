package repositories

import (
	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ GroceryEntryRepository = &GroceryEntryRepositoryImpl{}

type GroceryEntryRepository interface {
	GetByQuery(q *models.GroceryEntry) (*models.GroceryEntry, error)
	FindByQuery(q *models.GroceryEntry) ([]models.GroceryEntry, error)
	FindByGuildID(guildID string) ([]models.GroceryEntry, error)
}

type GroceryEntryRepositoryImpl struct {
	DB *gorm.DB
}

func (r *GroceryEntryRepositoryImpl) GetByQuery(q *models.GroceryEntry) (*models.GroceryEntry, error) {
	g := models.GroceryEntry{}
	if res := r.DB.Where(q).Take(g); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return &g, nil
}

func (r *GroceryEntryRepositoryImpl) FindByQuery(q *models.GroceryEntry) ([]models.GroceryEntry, error) {
	entries := make([]models.GroceryEntry, 0)
	if res := r.DB.Where(q).Find(&entries); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return entries, nil
}

func (r *GroceryEntryRepositoryImpl) FindByGuildID(guildID string) ([]models.GroceryEntry, error) {
	return r.FindByQuery(&models.GroceryEntry{GuildID: guildID})
}
