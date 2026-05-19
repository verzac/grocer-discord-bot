package repositories

import (
	"context"
	"errors"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ WaitlistIosRepository = &WaitlistIosRepositoryImpl{}

type WaitlistIosRepository interface {
	Create(ctx context.Context, entry *models.WaitlistIos) error
	FindByDiscordUserID(ctx context.Context, discordUserID string) (*models.WaitlistIos, error)
	UpdateByDiscordUserID(ctx context.Context, discordUserID, email, name string) error
}

type WaitlistIosRepositoryImpl struct {
	DB *gorm.DB
}

func (r *WaitlistIosRepositoryImpl) Create(ctx context.Context, entry *models.WaitlistIos) error {
	return r.DB.WithContext(ctx).Create(entry).Error
}

func (r *WaitlistIosRepositoryImpl) FindByDiscordUserID(ctx context.Context, discordUserID string) (*models.WaitlistIos, error) {
	var row models.WaitlistIos
	if err := r.DB.WithContext(ctx).Where("discord_user_id = ?", discordUserID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *WaitlistIosRepositoryImpl) UpdateByDiscordUserID(ctx context.Context, discordUserID, email, name string) error {
	return r.DB.WithContext(ctx).Model(&models.WaitlistIos{}).
		Where("discord_user_id = ?", discordUserID).
		Updates(map[string]interface{}{
			"email": email,
			"name":  name,
		}).Error
}
