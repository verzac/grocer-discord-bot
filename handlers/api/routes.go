package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/config"
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/handlers"
	apimw "github.com/verzac/grocer-discord-bot/handlers/api/middleware"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/monitoring/groprometheus"
	"github.com/verzac/grocer-discord-bot/repositories"
	"github.com/verzac/grocer-discord-bot/services/grocery"
	"github.com/verzac/grocer-discord-bot/services/registration"
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

// RegisterAndStart starts the API handler goroutine
func RegisterAndStart(logger *zap.Logger, db *gorm.DB, grobotVersion string, discordSess *discordgo.Session) error {
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
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, apimw.HeaderXGuildID},
	}))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 10 * time.Second,
	}))
	e.Use(middleware.Recover())
	e.Validator = utils.NewCustomValidator()

	groceryEntryRepo = &repositories.GroceryEntryRepositoryImpl{DB: db}
	groceryListRepo = &repositories.GroceryListRepositoryImpl{DB: db}
	guildRegistrationRepo = &repositories.GuildRegistrationRepositoryImpl{DB: db}
	apiClientRepo = &repositories.ApiClientRepositoryImpl{DB: db}

	e.Use(apimw.AuthMiddleware(apiClientRepo, logger, grobotVersion, discordSess))

	if grobotVersion == config.GrobotVersionLocal {
		e.POST("/.test/issue-jwt", func(c echo.Context) error {
			ctx := c.Request().Context()
			if auth.DefaultJWTIssuer == nil {
				return echo.NewHTTPError(500, "JWT issuer is not ready.")
			}
			discordUserID := "sub123"
			forParam := c.QueryParam("for")
			if forParam != "" {
				discordUserID = forParam
			}
			tokenStr, err := auth.DefaultJWTIssuer.Issue(ctx, discordUserID)
			if err != nil {
				return echo.NewHTTPError(500, err.Error())
			}
			return c.JSON(http.StatusOK, map[string]string{
				"access_token": tokenStr,
				"sub":          discordUserID,
			})
		})
	}

	e.GET("/metrics", groprometheus.PrometheusHandler())
	e.GET("/grocery-lists", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*apimw.AuthContext)
		guildID := authContext.GuildID
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
	e.DELETE("/groceries/:id", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*apimw.AuthContext)
		ctx := c.Request().Context()
		guildID := authContext.GuildID

		// Parse ID from path parameter
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(400, "Invalid ID format.")
		}

		// Look up the grocery entry by ID and GuildID
		entries, err := groceryEntryRepo.FindByQuery(&models.GroceryEntry{
			ID:      uint(id),
			GuildID: guildID,
		})
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			return echo.NewHTTPError(404, "Grocery entry not found.")
		}
		entry := entries[0]

		// Resolve grocery list if the entry has one
		var groceryList *models.GroceryList
		if entry.GroceryListID != nil && *entry.GroceryListID != 0 {
			groceryList, err = groceryListRepo.GetByQuery(&models.GroceryList{
				ID:      *entry.GroceryListID,
				GuildID: guildID,
			})
			if err != nil {
				return err
			}
		}

		// Delete the entry
		if err := groceryEntryRepo.WithContext(ctx).Delete(ctx, &entry); err != nil {
			return err
		}

		// Call post-deletion hook
		if err := grocery.Service.OnGroceryListEdit(ctx, groceryList, guildID); err != nil {
			logger.Error("Failed to run OnGroceryListEdit", zap.Error(err))
		}

		return c.NoContent(204)
	})
	// create new grocery
	e.POST("/groceries", func(c echo.Context) error {
		authContext := c.(*apimw.AuthContext)
		ctx := c.Request().Context()
		guildID := authContext.GuildID
		registrationContext, err := registration.Service.GetRegistrationContext(guildID)
		if err != nil {
			return err
		}
		groceryEntry := models.GroceryEntry{}
		if err := c.Bind(&groceryEntry); err != nil {
			return err
		}
		if err := c.Validate(&groceryEntry); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}
		groceryEntry.CreatedAt = time.Time{}
		groceryEntry.UpdatedAt = time.Time{}
		if groceryEntry.ID != 0 {
			return echo.NewHTTPError(400, "ID must be empty.")
		}
		var groceryList *models.GroceryList
		if groceryEntry.GroceryListID != nil && *groceryEntry.GroceryListID != 0 {
			groceryList, err := groceryListRepo.GetByQuery(&models.GroceryList{
				ID:      *groceryEntry.GroceryListID,
				GuildID: guildID,
			})
			if err != nil {
				return err
			}
			if groceryList == nil {
				return echo.NewHTTPError(400, "Cannot find grocery list with that ID.")
			}
		}
		limitOk, groceryEntryLimit, err := grocery.Service.ValidateGroceryEntryLimit(ctx, registrationContext, guildID, 1)
		if err != nil {
			return err
		}
		if !limitOk {
			return echo.NewHTTPError(400, fmt.Sprintf("You've reached the max number of grocery entries that you can have for your server. Limit: %d | Server ID: %s", groceryEntryLimit, guildID))
		}
		inputEntries := []models.GroceryEntry{groceryEntry}
		rErr := groceryEntryRepo.WithContext(ctx).AddToGroceryList(groceryList, inputEntries, guildID)
		if rErr != nil {
			switch rErr.ErrCode {
			case repositories.ErrCodeValidationError:
				return echo.NewHTTPError(400, rErr.Error())
			default:
				return rErr
			}
		}
		if err := grocery.Service.OnGroceryListEdit(ctx, groceryList, guildID); err != nil {
			logger.Error("Failed to run OnGroceryListEdit", zap.Error(err))
		}
		return c.JSON(201, inputEntries[0])
	})
	e.GET("/registrations", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*apimw.AuthContext)
		guildID := authContext.GuildID
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
