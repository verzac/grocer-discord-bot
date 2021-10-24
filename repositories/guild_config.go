package repositories

import (
	"errors"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ GuildConfigRepository = &GuildConfigRepositoryImpl{}

type GuildConfigRepository interface {
	Get(guildID string) (*models.GuildConfig, error)
}

type GuildConfigRepositoryImpl struct {
	db *gorm.DB
}

func (r *GuildConfigRepositoryImpl) Get(guildID string) (guildConfig *models.GuildConfig, err error) {
	if res := r.db.Where(&models.GuildConfig{GuildID: guildID}).Take(guildConfig); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return
}
