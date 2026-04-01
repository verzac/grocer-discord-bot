package repositories

import (
	"context"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ UserSessionRepository = &UserSessionRepositoryImpl{}

type UserSessionRepository interface {
	UpsertByDiscordUserID(ctx context.Context, session *models.UserSession) error
	GetByDiscordUserID(ctx context.Context, discordUserID string) (*models.UserSession, error)
	GetByRefreshTokenHash(ctx context.Context, hash string) (*models.UserSession, error)
	DeleteByDiscordUserID(ctx context.Context, discordUserID string) error
}

type UserSessionRepositoryImpl struct {
	DB *gorm.DB
}

func (r *UserSessionRepositoryImpl) UpsertByDiscordUserID(ctx context.Context, session *models.UserSession) error {
	var existing models.UserSession
	err := r.DB.WithContext(ctx).Unscoped().Where("discord_user_id = ?", session.DiscordUserID).Take(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.WithContext(ctx).Create(session).Error
	}
	if err != nil {
		return err
	}
	session.ID = existing.ID
	session.CreatedAt = existing.CreatedAt
	session.DeletedAt = gorm.DeletedAt{}
	return r.DB.WithContext(ctx).Unscoped().Save(session).Error
}

func (r *UserSessionRepositoryImpl) GetByDiscordUserID(ctx context.Context, discordUserID string) (*models.UserSession, error) {
	var out models.UserSession
	err := r.DB.WithContext(ctx).Where("discord_user_id = ?", discordUserID).Take(&out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *UserSessionRepositoryImpl) GetByRefreshTokenHash(ctx context.Context, hash string) (*models.UserSession, error) {
	var out models.UserSession
	err := r.DB.WithContext(ctx).Where("refresh_token_hash = ?", hash).Take(&out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *UserSessionRepositoryImpl) DeleteByDiscordUserID(ctx context.Context, discordUserID string) error {
	return r.DB.WithContext(ctx).Where("discord_user_id = ?", discordUserID).Delete(&models.UserSession{}).Error
}
