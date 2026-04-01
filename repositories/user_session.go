package repositories

import (
	"context"
	"errors"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ UserSessionRepository = &UserSessionRepositoryImpl{}

type UserSessionRepository interface {
	WithContext(ctx context.Context) UserSessionRepository
	Upsert(ctx context.Context, s *models.UserSession) error
	GetByDiscordUserID(ctx context.Context, discordUserID string) (*models.UserSession, error)
	GetByGroRefreshTokenHash(ctx context.Context, hash string) (*models.UserSession, error)
}

type UserSessionRepositoryImpl struct {
	DB *gorm.DB
}

func (r *UserSessionRepositoryImpl) WithContext(ctx context.Context) UserSessionRepository {
	return &UserSessionRepositoryImpl{DB: r.DB.WithContext(ctx)}
}

func (r *UserSessionRepositoryImpl) Upsert(ctx context.Context, s *models.UserSession) error {
	return r.DB.WithContext(ctx).Save(s).Error
}

func (r *UserSessionRepositoryImpl) GetByDiscordUserID(ctx context.Context, discordUserID string) (*models.UserSession, error) {
	var out models.UserSession
	if err := r.DB.WithContext(ctx).Where("discord_user_id = ?", discordUserID).Take(&out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *UserSessionRepositoryImpl) GetByGroRefreshTokenHash(ctx context.Context, hash string) (*models.UserSession, error) {
	var out models.UserSession
	if err := r.DB.WithContext(ctx).Where("gro_refresh_token_hash = ?", hash).Take(&out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}
