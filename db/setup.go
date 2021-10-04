package db

import (
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	sqliteMigrate "github.com/golang-migrate/migrate/v4/database/sqlite"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// loads the file:// driver for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(gormDB *gorm.DB, zlogger *zap.Logger, botVersion string) error {
	sourceURL := os.Getenv("GROCER_BOT_DB_SOURCE_CHANGELOG")
	if sourceURL == "" {
		sourceURL = "file://db/changelog"
	}
	if strings.HasPrefix(sourceURL, "github://") && !strings.Contains(sourceURL, "#") {
		sourceURL += "#" + botVersion
	}
	db, err := gormDB.DB()
	if err != nil {
		return err
	}
	driver, err := sqliteMigrate.WithInstance(db, &sqliteMigrate.Config{})
	if err != nil {
		return err
	}
	zlogger.Info("Running golang-migrate...", zap.String("sourceURL", sourceURL))
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "gorm.db", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func Setup(dsn string, zlogger *zap.Logger, botVersion string) *gorm.DB {
	isDBDebugMode := os.Getenv("GROCER_BOT_DB_DEBUG")
	logLevel := logger.Error
	if isDBDebugMode == "true" {
		logLevel = logger.Info
	}
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				LogLevel:                  logLevel, // Log level
				IgnoreRecordNotFoundError: true,     // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,    // Disable color
			},
		),
	})
	if err != nil {
		panic(err)
	}
	zlogger.Info("Migrating...")
	if os.Getenv("DISABLE_MIGRATION") == "true" {
		zlogger.Warn("WARNING: Migration is disabled as DISABLE_MIGRATION is set to true.")
		return db
	}
	if err := Migrate(db, zlogger, botVersion); err != nil {
		panic(err)
	}
	// if err := db.Migrator().AutoMigrate(&models.GroceryEntry{}); err != nil {
	// 	panic(err)
	// }
	// if err := db.Migrator().AutoMigrate(&models.GuildConfig{}); err != nil {
	// 	panic(err)
	// }
	// if err := db.Migrator().AutoMigrate(&models.GroceryList{}); err != nil {
	// 	panic(err)
	// }
	zlogger.Info("Migration done!")
	return db
}
