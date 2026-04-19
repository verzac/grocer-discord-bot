package middleware

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/patrickmn/go-cache"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/config"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	// HeaderXGuildID selects the target guild for Bearer-authenticated requests (see SPEC-001).
	HeaderXGuildID = "X-Guild-ID"
)

type AuthContext struct {
	echo.Context
	// UserID is set for Bearer auth (JWT sub); empty for Basic auth.
	UserID  string
	GuildID string
}

var bearerGuildOKCache = cache.New(60*time.Second, 2*time.Minute)

const (
	CtxKeyIdentifier = "sub"
)

var (
	errIncorrectToken             = echo.NewHTTPError(403, "Forbidden.")
	bearerPathSkipGuildIDCheckMap = map[string]bool{
		"/guilds":      false,
		"/auth/logout": false,
	}
	skipAuthForPathsMap = map[string]bool{
		"/metrics": true, // prometheus metrics endpoint
	}
)

func AuthMiddleware(apiKeyRepo repositories.ApiClientRepository, logger *zap.Logger, grobotVersion string, discordSess *discordgo.Session) echo.MiddlewareFunc {
	logger = logger.Named("middleware.auth")
	rateLimiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 0, ExpiresIn: 30 * time.Second},
	)
	rateLimitMiddleware := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		IdentifierExtractor: func(c echo.Context) (string, error) {
			if clientID, ok := c.Get(CtxKeyIdentifier).(string); ok {
				return clientID, nil
			} else {
				return c.RealIP(), nil
			}
		},
		Store: rateLimiterStore,
	})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, ok := skipAuthForPathsMap[c.Request().URL.Path]; ok {
				return next(c)
			}
			if strings.HasPrefix(c.Request().URL.Path, "/auth/") && c.Request().URL.Path != "/auth/logout" {
				return next(c)
			}
			if grobotVersion == config.GrobotVersionLocal && strings.HasPrefix(c.Request().URL.Path, "/.test/issue-jwt") {
				return next(c)
			}
			authHeader := c.Request().Header.Get("Authorization")
			headerType, headerValue, err := getHeaderTypeAndValue(authHeader)
			if err != nil {
				return echo.NewHTTPError(401, err.Error())
			}
			switch headerType {
			case HeaderTypeBasic:
				// basic
				decodedAuthHeader, err := base64.StdEncoding.DecodeString(headerValue)
				if err != nil {
					logger.Warn("basic auth: invalid base64 in Authorization header", zap.Error(err))
					return echo.NewHTTPError(401, "Malformed authentication.")
				}
				splitTokens := strings.Split(string(decodedAuthHeader), ":")
				if len(splitTokens) != 2 {
					return echo.NewHTTPError(401, "Malformed authentication.")
				}
				clientID := splitTokens[0]
				clientSecret := splitTokens[1]
				c.Set(CtxKeyIdentifier, clientID)

				return rateLimitMiddleware(func(c echo.Context) error {
					apiClient, err := apiKeyRepo.GetApiClient(&models.ApiClient{
						ClientID: clientID,
					})
					if err != nil {
						return err
					}
					if apiClient == nil {
						logger.Debug("Cannot find API client in DB.", zap.String(CtxKeyIdentifier, clientID))
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
						GuildID: guildID,
					})
				})(c)
			case HeaderTypeBearer:
				ctx := c.Request().Context()
				if auth.DefaultJWTIssuer == nil {
					return echo.NewHTTPError(401, "Bearer authentication is not available.")
				}
				userInfo, err := auth.DefaultJWTIssuer.Verify(ctx, headerValue)
				if err != nil {
					return echo.NewHTTPError(401, "Invalid token.")
				}
				logger.Debug("got discord user id", zap.String("DiscordUserID", userInfo.DiscordUserID))
				discordUserID := userInfo.DiscordUserID
				c.Set(CtxKeyIdentifier, discordUserID)

				return rateLimitMiddleware(func(c echo.Context) error {
					path := c.Request().URL.Path
					if _, ok := bearerPathSkipGuildIDCheckMap[path]; ok {
						return next(&AuthContext{
							Context: c,
							UserID:  discordUserID,
							GuildID: "",
						})
					}
					guildID := strings.TrimSpace(c.Request().Header.Get(HeaderXGuildID))
					if guildID == "" {
						return echo.NewHTTPError(400, "X-Guild-ID header is required.")
					}
					if discordSess == nil {
						return echo.NewHTTPError(500, "Cannot verify token.")
					}
					cacheKey := discordUserID + ":" + guildID
					if _, ok := bearerGuildOKCache.Get(cacheKey); !ok {
						if _, err := discordSess.Guild(guildID); err != nil {
							logger.Debug("bearer auth: guild lookup failed (bot may not be in guild)", zap.Error(err))
							return errIncorrectToken
						}
						if _, err := discordSess.GuildMember(guildID, discordUserID); err != nil {
							logger.Debug("bearer auth: user is not a member of guild", zap.Error(err))
							return errIncorrectToken
						}
						bearerGuildOKCache.Set(cacheKey, true, cache.DefaultExpiration)
					}
					return next(&AuthContext{
						Context: c,
						UserID:  discordUserID,
						GuildID: guildID,
					})
				})(c)
			default:
				return echo.NewHTTPError(401, "Unsupported authentication type.")
			}
		}
	}
}
