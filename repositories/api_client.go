package repositories

import (
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ ApiClientRepository = &ApiClientRepositoryImpl{}

type ApiClientRepository interface {
	FindApiClientsByGuildID(guildID string) ([]models.ApiClient, error)
	Put(apiKey *models.ApiClient) error
	GetApiClient(q *models.ApiClient) (*models.ApiClient, error)
	DeleteAllByClientID(clientID string) error
}

// get the api key
// make the api key
// validate the api key
// check for owner role

type ApiClientRepositoryImpl struct {
	DB *gorm.DB
}

const guildScopePrefix = "guild:"

func (r *ApiClientRepositoryImpl) FindApiClientsByGuildID(guildID string) ([]models.ApiClient, error) {
	q := &models.ApiClient{
		Scope:     auth.GetScopeForGuild(guildID),
		DeletedAt: gorm.DeletedAt{Valid: false},
	}
	keys := make([]models.ApiClient, 0)
	if res := r.DB.Where(q).Find(&keys); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return keys, nil
		}
		return nil, res.Error
	}
	return keys, nil
}

func (r *ApiClientRepositoryImpl) Put(apiKey *models.ApiClient) error {
	res := r.DB.Save(apiKey)
	return res.Error
}

func (r *ApiClientRepositoryImpl) GetApiClient(q *models.ApiClient) (*models.ApiClient, error) {
	out := models.ApiClient{}
	if res := r.DB.Where(q).Take(&out); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return &out, nil
}

func (r *ApiClientRepositoryImpl) DeleteAllByClientID(clientID string) error {
	if res := r.DB.Delete(&models.ApiClient{}, "client_id = ?", clientID); res.Error != nil {
		return res.Error
	}
	return nil
}
