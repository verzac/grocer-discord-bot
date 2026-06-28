package routegrocerylists

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/handlers"
	apimw "github.com/verzac/grocer-discord-bot/handlers/api/middleware"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"github.com/verzac/grocer-discord-bot/services/grocery"
	"github.com/verzac/grocer-discord-bot/services/registration"
	"github.com/verzac/grocer-discord-bot/utils"
	"go.uber.org/zap"
)

// Register mounts /grocery-lists mutation routes (POST, DELETE /:id, PATCH /:id).
// Not yet implemented — see handlers/grolist.go (newList, deleteList, editList).
func Register(
	e *echo.Echo,
	logger *zap.Logger,
	groceryListRepo repositories.GroceryListRepository,
	groceryEntryRepo repositories.GroceryEntryRepository,
	grohereRecordRepo repositories.GrohereRecordRepository,
	discordSess *discordgo.Session,
) {
	logger = logger.Named("grocery-lists")

	e.POST("/grocery-lists", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*apimw.AuthContext)
		guildID := authContext.GuildID

		req := dto.CreateGroceryListRequest{}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(400, "Invalid request body.")
		}
		if err := c.Validate(&req); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}

		if err := utils.ValidateSublistLabel(req.ListLabel); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}

		registrationContext, err := registration.Service.GetRegistrationContext(guildID)
		if err != nil {
			return err
		}
		ctx := c.Request().Context()
		existingCount, err := groceryListRepo.WithContext(ctx).Count(&models.GroceryList{GuildID: guildID})
		if err != nil {
			return err
		}
		if existingCount+1 >= int64(registrationContext.MaxGroceryListsPerServer) {
			return echo.NewHTTPError(400, fmt.Sprintf(
				"You've reached the max number of grocery lists for this server (%d).",
				registrationContext.MaxGroceryListsPerServer,
			))
		}

		var fancyName string
		if req.FancyName != nil {
			fancyName = *req.FancyName
		}

		newList, err := groceryListRepo.WithContext(ctx).CreateGroceryList(guildID, req.ListLabel, fancyName)
		if err != nil {
			if err == repositories.ErrGroceryListDuplicate {
				return echo.NewHTTPError(400, err.Error())
			}
			return err
		}
		if err := grocery.Service.UpdateGuildGrohere(ctx, guildID); err != nil {
			logger.Error("Failed to update grohere after grocery list creation", zap.Error(err))
		}
		return c.JSON(201, newList)
	})
	e.DELETE("/grocery-lists/:id", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*apimw.AuthContext)
		guildID := authContext.GuildID

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			return echo.NewHTTPError(400, "Invalid ID format.")
		}

		ctx := c.Request().Context()
		groceryList, err := groceryListRepo.WithContext(ctx).GetByQuery(&models.GroceryList{
			ID:      uint(id),
			GuildID: guildID,
		})
		if err != nil {
			return err
		}
		if groceryList == nil {
			return echo.NewHTTPError(404, repositories.ErrGroceryListNotFound.Error())
		}

		count, err := groceryEntryRepo.GetCount(&models.GroceryEntry{GuildID: guildID, GroceryListID: &groceryList.ID})
		if err != nil {
			return err
		}
		if count > 0 {
			return echo.NewHTTPError(409, fmt.Sprintf("Cannot delete: grocery list still has %d entries.", count))
		}

		grohereRecords, err := grohereRecordRepo.FindByQuery(&models.GrohereRecord{
			GuildID:       guildID,
			GroceryListID: &groceryList.ID,
		})
		if err != nil {
			return err
		}
		for i := range grohereRecords {
			record := &grohereRecords[i]
			if discordSess != nil {
				_, err := discordSess.ChannelMessageEdit(
					record.GrohereChannelID,
					record.GrohereMessageID,
					fmt.Sprintf(":shopping_cart: %s\n *:wave: This grocery list has been deleted. Type `!grohere:<your-new-grocery-list>` to get a self-updating message for your grocery list!*", groceryList.GetName()),
				)
				if err != nil {
					logger.Error("Failed to edit grohere message after grocery list deletion", zap.Error(err))
				}
			}
			if err := grohereRecordRepo.Delete(record); err != nil {
				return err
			}
		}

		if err := groceryListRepo.WithContext(ctx).Delete(groceryList); err != nil {
			if err == repositories.ErrGroceryListNotFound {
				return echo.NewHTTPError(404, err.Error())
			}
			return err
		}

		if err := grocery.Service.UpdateGuildGrohere(ctx, guildID); err != nil {
			logger.Error("Failed to update grohere after grocery list deletion", zap.Error(err))
		}

		return c.NoContent(204)
	})
	e.PATCH("/grocery-lists/:id", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*apimw.AuthContext)
		guildID := authContext.GuildID

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			return echo.NewHTTPError(400, "Invalid ID format.")
		}

		req := map[string]json.RawMessage{}
		if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
			return echo.NewHTTPError(400, "Invalid request body.")
		}
		rawFancyName, ok := req["fancy_name"]
		if !ok {
			return echo.NewHTTPError(400, "fancy_name is required.")
		}
		var fancyName *string
		if err := json.Unmarshal(rawFancyName, &fancyName); err != nil {
			return echo.NewHTTPError(400, "Invalid request body.")
		}

		ctx := c.Request().Context()
		groceryList, err := groceryListRepo.WithContext(ctx).GetByQuery(&models.GroceryList{
			ID:      uint(id),
			GuildID: guildID,
		})
		if err != nil {
			return err
		}
		if groceryList == nil {
			return echo.NewHTTPError(404, repositories.ErrGroceryListNotFound.Error())
		}

		groceryList.FancyName = fancyName
		if err := groceryListRepo.WithContext(ctx).Save(groceryList); err != nil {
			return err
		}

		if err := grocery.Service.OnGroceryListEdit(ctx, groceryList, guildID); err != nil {
			logger.Error("Failed to update grohere after grocery list edit", zap.Error(err))
		}

		return c.JSON(200, groceryList)
	})
}
