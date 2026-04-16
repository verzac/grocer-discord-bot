package repositories

import (
	"context"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ UserSessionRepository = &UserSessionRepositoryImpl{}

type UserSessionRepository interface {
	WithContext(ctx context.Context) UserSessionRepository
	CreateSession(ctx context.Context, session *models.UserSession) error
	FindByRefreshTokenHash(ctx context.Context, hash string) (*models.UserSession, error)
	UpdateSession(ctx context.Context, session *models.UserSession) error
	DeleteByDiscordUserID(ctx context.Context, discordUserID string) error
}

type UserSessionRepositoryImpl struct {
	DB *gorm.DB
}

func (r *UserSessionRepositoryImpl) WithContext(ctx context.Context) UserSessionRepository {
	return &UserSessionRepositoryImpl{DB: r.DB.WithContext(ctx)}
}

func (r *UserSessionRepositoryImpl) CreateSession(ctx context.Context, session *models.UserSession) error {
	return r.DB.WithContext(ctx).Create(session).Error
}

func (r *UserSessionRepositoryImpl) FindByRefreshTokenHash(ctx context.Context, hash string) (*models.UserSession, error) {
	var s models.UserSession
	if err := r.DB.WithContext(ctx).Where("refresh_token_hash = ?", hash).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *UserSessionRepositoryImpl) UpdateSession(ctx context.Context, session *models.UserSession) error {
	return r.DB.WithContext(ctx).Save(session).Error
}

func (r *UserSessionRepositoryImpl) DeleteByDiscordUserID(ctx context.Context, discordUserID string) error {
	return r.DB.WithContext(ctx).Where("discord_user_id = ?", discordUserID).Delete(&models.UserSession{}).Error
}
