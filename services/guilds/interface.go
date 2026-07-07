package guilds

import (
	"context"

	"github.com/verzac/grocer-discord-bot/dto"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Service GuildsService
)

type GuildsService interface {
	ResetGuild(ctx context.Context, guildID string) error
	GetUserGuilds(ctx context.Context, userID string) ([]dto.UserGuild, error)
}

type GuildsServiceImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func Init(db *gorm.DB, logger *zap.Logger) {
	if Service == nil {
		Service = &GuildsServiceImpl{db: db, logger: logger.Named("guilds")}
	}
}
