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
	_ "github.com/golang-migrate/migrate/v4/source/github"
)

const prefixGithub = "github://"
const unobfuscatedSecretCharCount = 8

func obfuscateSecret(sourceURL string) string {
	if !strings.HasPrefix(sourceURL, prefixGithub) {
		return sourceURL
	}
	shouldObfuscate := false
	buildString := ""
	obfuscatedCharCount := 0
	for _, r := range strings.TrimPrefix(sourceURL, prefixGithub) {
		buildStringChar := string(r)
		if buildStringChar == ":" {
			shouldObfuscate = true
		} else if buildStringChar == "@" {
			shouldObfuscate = false
		} else if shouldObfuscate {
			if obfuscatedCharCount > unobfuscatedSecretCharCount {
				buildStringChar = "*"
			} else {
				obfuscatedCharCount += 1
			}
		}
		buildString += buildStringChar
	}
	return prefixGithub + buildString
}

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
	zlogger.Info("Running golang-migrate...", zap.String("sourceURL", obfuscateSecret(sourceURL)))
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "gorm.db", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	version, dirty, err := m.Version()
	if err != nil {
		return err
	}
	zlogger.Info("Migration applied.", zap.Uint("DBVersion", version), zap.Bool("DBDirty", dirty))
	return nil
}

func Setup(dsn string, zlogger *zap.Logger, botVersion string) *gorm.DB {
	if _, err := os.Stat(dsn); err == nil && os.Getenv("GROCER_BOT_DB_DROP") == "true" && botVersion == "local" {
		zlogger.Info("WARN: Dropping DB.")
		if err := os.Remove(dsn); err != nil {
			zlogger.Error("Failed to remove gorm.db.", zap.Error(err))
		}
	}
	isDBDebugMode := os.Getenv("GROCER_BOT_DB_DEBUG")
	logLevel := logger.Error
	if isDBDebugMode == "true" {
		zlogger.Info("Using DB debug mode.")
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
