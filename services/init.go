package services

import (
	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/services/announcement"
	"github.com/verzac/grocer-discord-bot/services/grocery"
	"github.com/verzac/grocer-discord-bot/services/guildconfig"
	"github.com/verzac/grocer-discord-bot/services/registration"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func InitServices(db *gorm.DB, logger *zap.Logger, sess *discordgo.Session) {
	registration.Init(db, logger)
	grocery.Init(db, logger, sess)
	guildconfig.Init(db, logger)
	announcement.Init(db, logger)
}
