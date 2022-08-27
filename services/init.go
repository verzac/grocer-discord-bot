package services

import (
	"github.com/verzac/grocer-discord-bot/services/registration"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func InitServices(db *gorm.DB, logger *zap.Logger) {
	registration.Init(db, logger)
}
