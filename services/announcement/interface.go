package announcement

import (
	"context"

	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Service AnnouncementService
)

type AnnouncementService interface {
	AugmentMessageWithAnnouncement(ctx context.Context, guildID string, message string) (string, error)
}

type AnnouncementServiceImpl struct {
	guildconfigRepo repositories.GuildConfigRepository
	logger          *zap.Logger
}

func Init(db *gorm.DB, logger *zap.Logger) {
	if Service == nil {
		Service = &AnnouncementServiceImpl{
			guildconfigRepo: &repositories.GuildConfigRepositoryImpl{DB: db},
			logger:          logger.Named("announcement"),
		}
	}
}
