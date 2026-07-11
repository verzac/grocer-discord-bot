package routeauth

import (
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/verzac/grocer-discord-bot/services/demo"
	"go.uber.org/zap"
)

type demoLoginRequest struct {
	Password string `json:"password"`
}

func demoLoginRequestFields(c echo.Context) []zap.Field {
	return []zap.Field{
		zap.String("ip", c.RealIP()),
		zap.String("user_agent", c.Request().UserAgent()),
		zap.String("referer", c.Request().Header.Get("Referer")),
		zap.String("content_type", c.Request().Header.Get("Content-Type")),
		zap.String("x_forwarded_for", c.Request().Header.Get("X-Forwarded-For")),
	}
}

func RegisterDemoLogin(e *echo.Echo, logger *zap.Logger, demoService *demo.DemoService) {
	logger = logger.Named("auth.demo-login")

	rateLimiter := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: 1, Burst: 0, ExpiresIn: 30 * time.Second},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			fields := demoLoginRequestFields(c)
			fields = append(fields, zap.String("identifier", identifier), zap.Error(err))
			logger.Warn("demo login rate limited", fields...)
			return echo.NewHTTPError(429, "Too many requests.")
		},
	})

	e.POST("/auth/demo-login", func(c echo.Context) error {
		logger.Warn("demo login attempt", demoLoginRequestFields(c)...)

		ctx := c.Request().Context()
		var body demoLoginRequest
		if err := c.Bind(&body); err != nil {
			return echo.NewHTTPError(400, "Invalid request body.")
		}

		result, err := demoService.Login(ctx, body.Password)
		if err != nil {
			switch {
			case errors.Is(err, demo.ErrInvalidPassword):
				return echo.NewHTTPError(401, "Invalid demo password.")
			case errors.Is(err, demo.ErrNotSeeded):
				return echo.NewHTTPError(500, err.Error())
			default:
				return echo.NewHTTPError(500, "Cannot issue demo session.")
			}
		}

		return c.JSON(200, tokenResponse{
			AccessToken: result.AccessToken,
			// note: refresh token is not available for demo mode because demo should not be longer than 15 mins
			RefreshToken: "NOT AVAILABLE",
			ExpiresIn:    result.ExpiresIn,
		})
	}, rateLimiter)
}
