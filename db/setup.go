package db

import (
	"log"
	"os"

	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Setup(dsn string, zlogger *zap.Logger) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				LogLevel:                  logger.Error, // Log level
				IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,        // Disable color
			},
		),
	})
	if err != nil {
		panic(err)
	}
	zlogger.Info("Auto-Migrating...")
	if err := db.Migrator().AutoMigrate(&models.GroceryEntry{}); err != nil {
		panic(err)
	}
	if err := db.Migrator().AutoMigrate(&models.GuildConfig{}); err != nil {
		panic(err)
	}
	zlogger.Info("Migration done!")
	return db
}
