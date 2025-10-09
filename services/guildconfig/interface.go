package guildconfig

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Service GuildConfigService
)

type GuildConfigService interface {
	InitialiseGuildConfig(s *discordgo.Session)
}

type GuildConfigServiceImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func Init(db *gorm.DB, logger *zap.Logger) {
	if Service == nil {
		Service = &GuildConfigServiceImpl{
			db:     db,
			logger: logger.Named("guildconfig"),
		}
	}
}
