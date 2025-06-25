package api

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthContext struct {
	echo.Context
	Scope string
}

var (
	errIncorrectToken = echo.NewHTTPError(403, "Forbidden.")
)

func AuthMiddleware(apiKeyRepo repositories.ApiClientRepository, logger *zap.Logger) echo.MiddlewareFunc {
	logger = logger.Named("middleware.auth")
	rateLimiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 0, ExpiresIn: 30 * time.Second},
	)
	rateLimitMiddleware := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		IdentifierExtractor: func(c echo.Context) (string, error) {
			if clientID, ok := c.Get("clientID").(string); ok {
				return clientID, nil
			} else {
				return c.RealIP(), nil
			}
		},
		Store: rateLimiterStore,
	})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/metrics" {
				// skip auth for metrics endpoint
				return next(c)
			}
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
				return echo.NewHTTPError(401, "Missing authentication.")
			}
			authHeaderTrimmed := strings.TrimPrefix(authHeader, "Basic ")
			decodedAuthHeader, err := base64.StdEncoding.DecodeString(authHeaderTrimmed)
			if err != nil {
				return err
			}
			splitTokens := strings.Split(string(decodedAuthHeader), ":")
			if len(splitTokens) != 2 {
				return echo.NewHTTPError(401, "Malformed authentication.")
			}
			clientID := splitTokens[0]
			clientSecret := splitTokens[1]
			c.Set("clientID", clientID)

			return rateLimitMiddleware(func(c echo.Context) error {
				apiClient, err := apiKeyRepo.GetApiClient(&models.ApiClient{
					ClientID: clientID,
				})
				if err != nil {
					return err
				}
				if apiClient == nil {
					logger.Debug("Cannot find API client in DB.", zap.String("clientID", clientID))
					return errIncorrectToken
				}
				if err := bcrypt.CompareHashAndPassword([]byte(apiClient.ClientSecret), []byte(clientSecret)); err != nil {
					if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
						c.Logger().Error(err)
						return err
					}
					return errIncorrectToken
				}

				return next(&AuthContext{
					Context: c,
					Scope:   apiClient.Scope,
				})
			})(c)
		}
	}
}
