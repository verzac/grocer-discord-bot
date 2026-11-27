package repositories

import (
	"context"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ IntentTokenRepository = &IntentTokenRepositoryImpl{}

type IntentTokenRepository interface {
	// Upsert stores token as the user's only intent token, replacing any existing one in a single statement.
	Upsert(ctx context.Context, token *models.IntentToken) error
	FindByTokenHash(ctx context.Context, hash string) (*models.IntentToken, error)
	DeleteByDiscordUserID(ctx context.Context, discordUserID string) error
}

type IntentTokenRepositoryImpl struct {
	DB *gorm.DB
}

func (r *IntentTokenRepositoryImpl) Upsert(ctx context.Context, token *models.IntentToken) error {
	// Relies on the unique index on discord_user_id. On the update path SQLite does not refresh
	// last_insert_rowid, so token.ID may not reflect the stored row.
	return r.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "discord_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"token_hash", "expires_at", "created_at"}),
	}).Create(token).Error
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
