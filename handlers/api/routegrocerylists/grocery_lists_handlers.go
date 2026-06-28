package routegrocerylists

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/handlers"
	apimw "github.com/verzac/grocer-discord-bot/handlers/api/middleware"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
)

// Register mounts /grocery-lists mutation routes (POST, DELETE /:id, PATCH /:id).
// Not yet implemented — see handlers/grolist.go (newList, deleteList, editList).
func Register(
	e *echo.Echo,
	logger *zap.Logger,
	groceryListRepo repositories.GroceryListRepository,
	groceryEntryRepo repositories.GroceryEntryRepository,
) {
	logger = logger.Named("grocery-lists")

	e.POST("/grocery-lists", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*apimw.AuthContext)
		_ = authContext.GuildID

		req := dto.CreateGroceryListRequest{}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(400, "Invalid request body.")
		}
		if err := c.Validate(&req); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}

		return echo.NewHTTPError(501, "Not implemented.")
	})
	e.DELETE("/grocery-lists/:id", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*apimw.AuthContext)
		_ = authContext.GuildID

		idStr := c.Param("id")
		if _, err := strconv.ParseUint(idStr, 10, 64); err != nil {
			return echo.NewHTTPError(400, "Invalid ID format.")
		}

		return echo.NewHTTPError(501, "Not implemented.")
	})
	e.PATCH("/grocery-lists/:id", func(c echo.Context) error {
		defer handlers.Recover(logger)
		authContext := c.(*apimw.AuthContext)
		_ = authContext.GuildID

		idStr := c.Param("id")
		if _, err := strconv.ParseUint(idStr, 10, 64); err != nil {
			return echo.NewHTTPError(400, "Invalid ID format.")
		}

		req := dto.UpdateGroceryListRequest{}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(400, "Invalid request body.")
		}
		if err := c.Validate(&req); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}

		return echo.NewHTTPError(501, "Not implemented.")
	})
}
