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
	"github.com/patrickmn/go-cache"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const headerGuildID = "X-Guild-ID"

type AuthContext struct {
	echo.Context
	AuthScheme string // "basic" or "bearer"
	Scope      string // Basic: guild-scoped
	UserID     string // Bearer: Discord user ID from JWT
	GuildID    string // Basic: from scope; Bearer: from X-Guild-ID after membership check
}

// ResolveGuildID returns the target guild for CRUD: Basic uses api client scope; Bearer uses X-Guild-ID.
func (a *AuthContext) ResolveGuildID() string {
	if a.AuthScheme == "bearer" {
		return a.GuildID
	}
	return auth.GetGuildIDFromScope(a.Scope)
}

var (
	errIncorrectToken = echo.NewHTTPError(403, "Forbidden.")
)

var guildMembershipCache = cache.New(60*time.Second, 2*time.Minute)

func AuthMiddleware(
	apiKeyRepo repositories.ApiClientRepository,
	dg *discordgo.Session,
	logger *zap.Logger,
) echo.MiddlewareFunc {
	logger = logger.Named("middleware.auth")
	rateLimiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 0, ExpiresIn: 30 * time.Second},
	)
	rateLimitMiddleware := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		IdentifierExtractor: func(c echo.Context) (string, error) {
			if id, ok := c.Get("rateLimitKey").(string); ok && id != "" {
				return id, nil
			}
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
			p := c.Request().URL.Path
			// skip auth for metrics endpoint
			if p == "/metrics" {
				return next(c)
			}
			// unauthenticated OAuth initiation, callback, and refresh
			if p == "/auth/discord" || p == "/auth/discord/callback" ||
				(p == "/auth/refresh" && c.Request().Method == http.MethodPost) {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(401, "Missing authentication.")
			}

			if strings.HasPrefix(authHeader, "Basic ") {
				return handleBasicAuth(c, next, authHeader, apiKeyRepo, rateLimitMiddleware, logger)
			}
			if strings.HasPrefix(authHeader, "Bearer ") {
				return handleBearerAuth(c, next, authHeader, dg, rateLimitMiddleware, logger)
			}
			return echo.NewHTTPError(401, "Missing authentication.")
		}
	}
}

func handleBasicAuth(
	c echo.Context,
	next echo.HandlerFunc,
	authHeader string,
	apiKeyRepo repositories.ApiClientRepository,
	rateLimitMiddleware echo.MiddlewareFunc,
	logger *zap.Logger,
) error {
	authHeaderTrimmed := strings.TrimPrefix(authHeader, "Basic ")
	decodedAuthHeader, err := base64.StdEncoding.DecodeString(authHeaderTrimmed)
	if err != nil {
		return echo.NewHTTPError(401, "Malformed authentication.")
	}
	splitTokens := strings.Split(string(decodedAuthHeader), ":")
	if len(splitTokens) != 2 {
		return echo.NewHTTPError(401, "Malformed authentication.")
	}
	clientID := splitTokens[0]
	clientSecret := splitTokens[1]
	c.Set("clientID", clientID)
	c.Set("rateLimitKey", "basic:"+clientID)

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

		guildID := auth.GetGuildIDFromScope(apiClient.Scope)
		return next(&AuthContext{
			Context:    c,
			AuthScheme: "basic",
			Scope:      apiClient.Scope,
			GuildID:    guildID,
		})
	})(c)
}

func handleBearerAuth(
	c echo.Context,
	next echo.HandlerFunc,
	authHeader string,
	dg *discordgo.Session,
	rateLimitMiddleware echo.MiddlewareFunc,
	logger *zap.Logger,
) error {
	if !oauthEnvComplete() {
		return echo.NewHTTPError(401, "Bearer authentication is not configured.")
	}
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		return echo.NewHTTPError(401, "Missing bearer token.")
	}
	claims, err := parseAccessJWT(tokenStr)
	if err != nil {
		return jwtHTTPError(err)
	}
	if claims.Subject == "" {
		return echo.NewHTTPError(401, "Invalid token subject.")
	}
	c.Set("rateLimitKey", "bearer:"+claims.Subject)

	path := c.Request().URL.Path
	if path == "/guilds" || strings.HasPrefix(path, "/auth/") {
		return rateLimitMiddleware(func(c echo.Context) error {
			return next(&AuthContext{
				Context:    c,
				AuthScheme: "bearer",
				UserID:     claims.Subject,
			})
		})(c)
	}

	guildID := strings.TrimSpace(c.Request().Header.Get(headerGuildID))
	if guildID == "" {
		return echo.NewHTTPError(400, "X-Guild-ID header is required for Bearer authentication.")
	}

	return rateLimitMiddleware(func(c echo.Context) error {
		ok, err := verifyUserGuildMembership(dg, claims.Subject, guildID, logger)
		if err != nil {
			logger.Warn("Guild membership check failed", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadGateway, "Could not verify guild membership.")
		}
		if !ok {
			return echo.NewHTTPError(403, "Not a member of this guild or bot is not installed there.")
		}
		return next(&AuthContext{
			Context:    c,
			AuthScheme: "bearer",
			UserID:     claims.Subject,
			GuildID:    guildID,
		})
	})(c)
}

func verifyUserGuildMembership(s *discordgo.Session, userID, guildID string, logger *zap.Logger) (bool, error) {
	cacheKey := userID + "|" + guildID
	if v, found := guildMembershipCache.Get(cacheKey); found {
		return v.(bool), nil
	}

	if s == nil {
		return false, errors.New("discord session not available")
	}
	if _, err := s.State.Guild(guildID); err != nil {
		guildMembershipCache.Set(cacheKey, false, cache.DefaultExpiration)
		return false, nil
	}

	_, err := s.GuildMember(guildID, userID)
	if err != nil {
		var restErr *discordgo.RESTError
		if errors.As(err, &restErr) && restErr != nil && restErr.Response != nil &&
			restErr.Response.StatusCode == http.StatusNotFound {
			guildMembershipCache.Set(cacheKey, false, cache.DefaultExpiration)
			return false, nil
		}
		return false, err
	}
	guildMembershipCache.Set(cacheKey, true, cache.DefaultExpiration)
	return true, nil
}
