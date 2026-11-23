package repositories

import (
	"context"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ IntentTokenRepository = &IntentTokenRepositoryImpl{}

type IntentTokenRepository interface {
	WithContext(ctx context.Context) IntentTokenRepository
	Create(ctx context.Context, token *models.IntentToken) error
	FindByTokenHash(ctx context.Context, hash string) (*models.IntentToken, error)
	DeleteByDiscordUserID(ctx context.Context, discordUserID string) error
}

type IntentTokenRepositoryImpl struct {
	DB *gorm.DB
}

func (r *IntentTokenRepositoryImpl) WithContext(ctx context.Context) IntentTokenRepository {
	return &IntentTokenRepositoryImpl{DB: r.DB.WithContext(ctx)}
}

func (r *IntentTokenRepositoryImpl) Create(ctx context.Context, token *models.IntentToken) error {
	return r.DB.WithContext(ctx).Create(token).Error
}

func (r *IntentTokenRepositoryImpl) FindByTokenHash(ctx context.Context, hash string) (*models.IntentToken, error) {
	var t models.IntentToken
	if err := r.DB.WithContext(ctx).Where("token_hash = ?", hash).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *IntentTokenRepositoryImpl) DeleteByDiscordUserID(ctx context.Context, discordUserID string) error {
	return r.DB.WithContext(ctx).Where("discord_user_id = ?", discordUserID).Delete(&models.IntentToken{}).Error
}
