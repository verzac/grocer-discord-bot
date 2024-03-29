package repositories

import (
	"errors"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ GuildConfigRepository = &GuildConfigRepositoryImpl{}

type GuildConfigRepository interface {
	Get(guildID string) (*models.GuildConfig, error)
	Put(g *models.GuildConfig) error
}

type GuildConfigRepositoryImpl struct {
	DB *gorm.DB
}

func (r *GuildConfigRepositoryImpl) Get(guildID string) (guildConfig *models.GuildConfig, err error) {
	guildConfig = &models.GuildConfig{}
	if res := r.DB.Where(&models.GuildConfig{GuildID: guildID}).Take(guildConfig); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return
}

func (r *GuildConfigRepositoryImpl) Put(g *models.GuildConfig) error {
	if r := r.DB.Save(g); r.Error != nil {
		return r.Error
	}
	return nil
}
