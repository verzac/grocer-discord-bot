package routeauth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
)

const discordUsersMeURL = "https://discord.com/api/users/@me"

type discordUserMe struct {
	ID string `json:"id"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register mounts Discord OAuth and refresh-token routes on e.
func Register(
	e *echo.Echo,
	logger *zap.Logger,
	oauthSetup *auth.OAuthSetup,
	userSessionRepo repositories.UserSessionRepository,
) {
	logger = logger.Named("auth")
	stateCache := cache.New(5*time.Minute, 10*time.Minute)

	e.GET("/auth/discord", func(c echo.Context) error {
		state, err := auth.GenerateKey()
		if err != nil {
			logger.Error("oauth state generation failed", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot start login.")
		}
		stateCache.Set(state, true, cache.DefaultExpiration)
		authURL := oauthSetup.OAuth2.AuthCodeURL(state)
		return c.Redirect(302, authURL)
	})

	e.GET("/auth/discord/callback", func(c echo.Context) error {
		ctx := c.Request().Context()
		if qErr := c.QueryParam("error"); qErr != "" {
			desc := c.QueryParam("error_description")
			logger.Warn("discord oauth error", zap.String("error", qErr), zap.String("error_description", desc))
			return echo.NewHTTPError(401, "Discord authorization was denied or failed.")
		}
		code := c.QueryParam("code")
		state := c.QueryParam("state")
		if code == "" || state == "" {
			return echo.NewHTTPError(400, "Missing code or state.")
		}
		if _, ok := stateCache.Get(state); !ok {
			return echo.NewHTTPError(400, "Invalid or expired state.")
		}
		stateCache.Delete(state)

		token, err := oauthSetup.OAuth2.Exchange(ctx, code)
		if err != nil {
			logger.Error("discord token exchange failed", zap.Error(err))
			return echo.NewHTTPError(502, "Could not complete login with Discord.")
		}

		client := oauthSetup.OAuth2.Client(ctx, token)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, discordUsersMeURL, nil)
		if err != nil {
			logger.Error("build @me request", zap.Error(err))
			return echo.NewHTTPError(502, "Could not reach Discord.")
		}
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("discord @me request failed", zap.Error(err))
			return echo.NewHTTPError(502, "Could not reach Discord.")
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logger.Warn("discord @me non-OK", zap.Int("status", resp.StatusCode))
			return echo.NewHTTPError(502, "Could not verify Discord account.")
		}
		var du discordUserMe
		if err := json.NewDecoder(resp.Body).Decode(&du); err != nil {
			logger.Error("decode discord user", zap.Error(err))
			return echo.NewHTTPError(502, "Could not verify Discord account.")
		}
		if du.ID == "" {
			return echo.NewHTTPError(502, "Could not verify Discord account.")
		}

		if auth.DefaultJWTIssuer == nil {
			return echo.NewHTTPError(500, "JWT issuer is not ready.")
		}
		accessJWT, err := auth.DefaultJWTIssuer.Issue(ctx, du.ID)
		if err != nil {
			logger.Error("issue access jwt", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot issue session.")
		}
		refreshPlain, refreshHash, err := auth.GenerateRefreshToken()
		if err != nil {
			logger.Error("issue refresh token", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot issue session.")
		}
		expiry := time.Now().UTC().Add(auth.DefaultRefreshTokenTTL)

		if err := userSessionRepo.WithContext(ctx).DeleteByDiscordUserID(ctx, du.ID); err != nil {
			logger.Error("delete prior user sessions", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot issue session.")
		}
		logger.Debug("received user info", zap.Any("userInfo", du))
		sess := &models.UserSession{
			DiscordUserID:      du.ID,
			RefreshTokenHash:   refreshHash,
			RefreshTokenExpiry: expiry,
		}
		if err := userSessionRepo.WithContext(ctx).CreateSession(ctx, sess); err != nil {
			logger.Error("save user session", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot issue session.")
		}

		u, err := url.Parse(oauthSetup.AppRedirectURI)
		if err != nil {
			logger.Error("parse app redirect uri", zap.Error(err))
			return echo.NewHTTPError(500, "Invalid app redirect configuration.")
		}
		q := u.Query()
		q.Set("access_token", accessJWT)
		q.Set("refresh_token", refreshPlain)
		u.RawQuery = q.Encode()
		return c.Redirect(302, u.String())
	})

	e.POST("/auth/refresh", func(c echo.Context) error {
		ctx := c.Request().Context()
		var body refreshTokenRequest
		if err := c.Bind(&body); err != nil {
			return echo.NewHTTPError(400, "Invalid request body.")
		}
		if body.RefreshToken == "" {
			return echo.NewHTTPError(401, "Refresh token is required.")
		}
		hash := auth.HashRefreshToken(body.RefreshToken)
		sess, err := userSessionRepo.WithContext(ctx).FindByRefreshTokenHash(ctx, hash)
		if err != nil {
			logger.Error("lookup refresh session", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot refresh session.")
		}
		if sess == nil || time.Now().UTC().After(sess.RefreshTokenExpiry) {
			return echo.NewHTTPError(401, "Refresh token is invalid or expired.")
		}
		if auth.DefaultJWTIssuer == nil {
			return echo.NewHTTPError(500, "JWT issuer is not ready.")
		}
		accessJWT, err := auth.DefaultJWTIssuer.Issue(ctx, sess.DiscordUserID)
		if err != nil {
			logger.Error("issue access jwt on refresh", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot refresh session.")
		}
		newPlain, newHash, err := auth.GenerateRefreshToken()
		if err != nil {
			logger.Error("rotate refresh token", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot refresh session.")
		}
		sess.RefreshTokenHash = newHash
		sess.RefreshTokenExpiry = time.Now().UTC().Add(auth.DefaultRefreshTokenTTL)
		if err := userSessionRepo.WithContext(ctx).UpdateSession(ctx, sess); err != nil {
			logger.Error("save rotated refresh token", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot refresh session.")
		}
		return c.JSON(http.StatusOK, refreshTokenResponse{
			AccessToken:  accessJWT,
			RefreshToken: newPlain,
		})
	})
}
