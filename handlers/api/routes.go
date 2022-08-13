package api

import (
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"github.com/verzac/grocer-discord-bot/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	groceryEntryRepo      repositories.GroceryEntryRepository
	guildRegistrationRepo repositories.GuildRegistrationRepository
	groceryListRepo       repositories.GroceryListRepository
	apiClientRepo         repositories.ApiClientRepository
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
	groceryListRepo = &repositories.GroceryListRepositoryImpl{DB: db}
	guildRegistrationRepo = &repositories.GuildRegistrationRepositoryImpl{DB: db}
	apiClientRepo = &repositories.ApiClientRepositoryImpl{DB: db}
	e.Use(AuthMiddleware(apiClientRepo, logger))
	e.GET("/grocery-lists", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*AuthContext)
		guildID := auth.GetGuildIDFromScope(authContext.Scope)
		groceryEntries, err := groceryEntryRepo.FindByQuery(&models.GroceryEntry{GuildID: guildID})
		if err != nil {
			return c.String(500, err.Error())
		}
		groceryLists, err := groceryListRepo.FindByQuery(&models.GroceryList{GuildID: guildID})
		if err != nil {
			return c.String(500, err.Error())
		}
		out := &dto.GuildGroceryList{
			GuildID:        guildID,
			GroceryEntries: groceryEntries,
			GroceryLists:   groceryLists,
		}
		return c.JSON(200, out)
	})
	e.GET("/registrations", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*AuthContext)
		guildID := auth.GetGuildIDFromScope(authContext.Scope)
		registrations, err := guildRegistrationRepo.FindByQuery(&models.GuildRegistration{GuildID: guildID})
		if err != nil {
			return c.String(500, err.Error())
		}
		return c.JSON(200, registrations)
	})
	// myUserID := "183947835467759617"
	// myUsername := "verzac"
	// myDiscriminator := "6377"
	// if err := registrationEntitlementRepo.Save(&models.RegistrationEntitlement{
	// 	// UserID: &myUserID,
	// 	Username:              &myUsername,
	// 	UsernameDiscriminator: &myDiscriminator,
	// 	MaxRedemption:         99,
	// 	RegistrationTierID:    1,
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
