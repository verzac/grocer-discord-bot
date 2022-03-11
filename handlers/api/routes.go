package api

import (
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"github.com/verzac/grocer-discord-bot/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	groceryEntryRepo            repositories.GroceryEntryRepository
	guildRegistrationRepo       repositories.GuildRegistrationRepository
	registrationEntitlementRepo repositories.RegistrationEntitlementRepository
)

// RegisterAndStart starts the API handler goroutine. TO BE USED IN DEV ONLY FOR NOW
func RegisterAndStart(logger *zap.Logger, db *gorm.DB) error {
	logger = logger.Named("api")
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())

	ao := os.Getenv("GROCER_BOT_API_ALLOW_ORIGINS")
	if ao == "" {
		ao = "http://localhost:3001,http://127.0.0.1:3001,http://localhost:3000,http://127.0.0.1:3000"
	}
	e.Logger.SetHeader("L-${time_rfc3339} ${level} ${short_file}:${line}")
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: strings.Split(ao, ","),
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 10 * time.Second,
	}))
	e.Validator = utils.NewCustomValidator()
	groceryEntryRepo = &repositories.GroceryEntryRepositoryImpl{DB: db}
	guildRegistrationRepo = &repositories.GuildRegistrationRepositoryImpl{DB: db}
	registrationEntitlementRepo = &repositories.RegistrationEntitlementRepositoryImpl{DB: db}
	e.GET("/grocery-lists/:guildID", func(c echo.Context) error {
		defer handlers.Recover(logger)
		guildID := c.Param("guildID")
		groceryEntries, err := groceryEntryRepo.FindByQuery(&models.GroceryEntry{GuildID: guildID})
		if err != nil {
			return c.String(500, err.Error())
		}
		return c.JSON(200, groceryEntries)
	})
	e.GET("/registrations/:guildID", func(c echo.Context) error {
		defer handlers.Recover(logger)
		guildID := c.Param("guildID")
		registrations, err := guildRegistrationRepo.FindByQuery(&models.GuildRegistration{GuildID: guildID})
		if err != nil {
			return c.String(500, err.Error())
		}
		return c.JSON(200, registrations)
	})
	// if err := registrationEntitlementRepo.Save(&models.RegistrationEntitlement{
	// 	UserID:             "183947835467759617",
	// 	MaxRedemption:      99,
	// 	RegistrationTierID: 1,
	// }); err != nil {
	// 	logger.Error("Failed to save entitlement.", zap.Error(err))
	// }
	// if err := guildRegistrationRepo.Save(&models.GuildRegistration{
	// 	GuildID:                       "815482602278354944",
	// 	RegistrationEntitlementUserID: "183947835467759617",
	// }); err != nil {
	// 	logger.Error("Failed to save registration.", zap.Error(err))
	// }
	logger.Info("Starting API!")
	return e.Start(":8080")
}
