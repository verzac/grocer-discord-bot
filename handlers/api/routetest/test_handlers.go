package routetest

import (
	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/auth"
	"go.uber.org/zap"
)

func Register(e *echo.Echo, logger *zap.Logger) {
	logger = logger.Named("test")
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
		return c.JSON(200, map[string]string{
			"access_token": tokenStr,
			"sub":          discordUserID,
		})
	})
}
