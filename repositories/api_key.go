package repositories

import (
	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ ApiKeyRepository = &ApiKeyRepositoryImpl{}

type ApiKeyRepository interface {
	FindApiKeysByGuildID(guildID string) ([]models.ApiKey, error)
	Put(apiKey *models.ApiKey) error
	GetScopeForGuild(guildID string) string
}

// get the api key
// make the api key
// validate the api key
// check for owner role

type ApiKeyRepositoryImpl struct {
	DB *gorm.DB
}

const guildScopePrefix = "guild:"

func (r *ApiKeyRepositoryImpl) FindApiKeysByGuildID(guildID string) ([]models.ApiKey, error) {
	q := &models.ApiKey{
		Scope: r.GetScopeForGuild(guildID),
	}
	keys := make([]models.ApiKey, 0)
	if res := r.DB.Where(q).Find(&keys); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return keys, nil
		}
		return nil, res.Error
	}
	return keys, nil
}

func (r *ApiKeyRepositoryImpl) Put(apiKey *models.ApiKey) error {
	res := r.DB.Save(apiKey)
	return res.Error
}

// GetScopeForGuild helper func to enforce consistent formatting
func (r *ApiKeyRepositoryImpl) GetScopeForGuild(guildID string) string {
	return guildScopePrefix + guildID
}
