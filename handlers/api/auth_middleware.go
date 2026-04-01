package api

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const headerGuildID = "X-Guild-ID"

type AuthContext struct {
	echo.Context
	Scope   string
	UserID  string
	GuildID string
}

type AuthMiddlewareDeps struct {
	ApiKeyRepo   repositories.ApiClientRepository
	Logger       *zap.Logger
	OAuthEnabled bool
	JWTKey       string
	BotSession   *discordgo.Session
}

var (
	errIncorrectToken = echo.NewHTTPError(403, "Forbidden.")
)

func pathRequiresGuildIDHeaderForEcho(c echo.Context) bool {
	switch c.Path() {
	case "/grocery-lists":
		return c.Request().Method == http.MethodGet
	case "/registrations":
		return c.Request().Method == http.MethodGet
	case "/groceries":
		return c.Request().Method == http.MethodPost
	case "/groceries/:id":
		return c.Request().Method == http.MethodDelete
	default:
		return false
	}
}

func AuthMiddleware(deps AuthMiddlewareDeps) echo.MiddlewareFunc {
	logger := deps.Logger.Named("middleware.auth")
	rateLimiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 0, ExpiresIn: 30 * time.Second},
	)
	rateLimitMiddleware := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		IdentifierExtractor: func(c echo.Context) (string, error) {
			if k, ok := c.Get("rateLimitKey").(string); ok && k != "" {
				return k, nil
			}
			if clientID, ok := c.Get("clientID").(string); ok {
				return clientID, nil
			}
			return c.RealIP(), nil
		},
		Store: rateLimiterStore,
	})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if c.Path() == "/guilds" && c.Request().Method == http.MethodGet {
				if !strings.HasPrefix(authHeader, "Bearer ") {
					return echo.NewHTTPError(http.StatusUnauthorized, "Bearer authentication required.")
				}
			}

			if strings.HasPrefix(authHeader, "Bearer ") {
				if !deps.OAuthEnabled || deps.JWTKey == "" {
					return echo.NewHTTPError(http.StatusUnauthorized, "OAuth2 is not configured.")
				}
				raw := strings.TrimPrefix(authHeader, "Bearer ")
				raw = strings.TrimSpace(raw)
				userID, err := parseAccessToken(deps.JWTKey, raw)
				if err != nil {
					if IsJWTExpired(err) {
						return echo.NewHTTPError(http.StatusUnauthorized, "Token expired.")
					}
					if IsJWTSignatureInvalid(err) {
						return echo.NewHTTPError(http.StatusForbidden, "Invalid token signature.")
					}
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token.")
				}
				c.Set("rateLimitKey", "user:"+userID)

				if c.Path() == "/guilds" && c.Request().Method == http.MethodGet {
					return rateLimitMiddleware(func(c echo.Context) error {
						return next(&AuthContext{
							Context: c,
							UserID:  userID,
						})
					})(c)
				}

				guildHeader := c.Request().Header.Get(headerGuildID)
				if pathRequiresGuildIDHeaderForEcho(c) {
					if guildHeader == "" {
						return echo.NewHTTPError(http.StatusBadRequest, "Missing "+headerGuildID+" header.")
					}
					if err := verifyUserGuildMembership(deps.BotSession, guildHeader, userID); err != nil {
						return err
					}
				}

				return rateLimitMiddleware(func(c echo.Context) error {
					ac := &AuthContext{
						Context: c,
						UserID:  userID,
						GuildID: guildHeader,
					}
					return next(ac)
				})(c)
			}

			if !strings.HasPrefix(authHeader, "Basic ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing authentication.")
			}
			authHeaderTrimmed := strings.TrimPrefix(authHeader, "Basic ")
			decodedAuthHeader, err := base64.StdEncoding.DecodeString(authHeaderTrimmed)
			if err != nil {
				return err
			}
			splitTokens := strings.Split(string(decodedAuthHeader), ":")
			if len(splitTokens) != 2 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Malformed authentication.")
			}
			clientID := splitTokens[0]
			clientSecret := splitTokens[1]
			c.Set("clientID", clientID)

			return rateLimitMiddleware(func(c echo.Context) error {
				apiClient, err := deps.ApiKeyRepo.GetApiClient(&models.ApiClient{
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
				guildID := auth.GetGuildIDFromScope(apiClient.Scope)
				return next(&AuthContext{
					Context: c,
					Scope:   apiClient.Scope,
					GuildID: guildID,
				})
			})(c)
		}
	}
}
