package routeauth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/auth"
	apimw "github.com/verzac/grocer-discord-bot/handlers/api/middleware"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const discordUsersMeURL = "https://discord.com/api/users/@me"

type discordUserMe struct {
	ID string `json:"id"`
}

type tokenExchangeRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Register mounts Discord OAuth and refresh-token routes on e.
func Register(
	e *echo.Echo,
	logger *zap.Logger,
	oauthSetup *auth.OAuthSetup,
	userSessionRepo repositories.UserSessionRepository,
) {
	logger = logger.Named("auth")

	e.POST("/auth/token", func(c echo.Context) error {
		ctx := c.Request().Context()
		var body tokenExchangeRequest
		if err := c.Bind(&body); err != nil {
			return echo.NewHTTPError(400, "Invalid request body.")
		}
		if body.Code == "" || body.CodeVerifier == "" || body.RedirectURI == "" {
			return echo.NewHTTPError(400, "code, code_verifier, and redirect_uri are required.")
		}
		if !oauthSetup.ValidateRedirectURI(body.RedirectURI) {
			return echo.NewHTTPError(400, "redirect_uri is not allowed.")
		}

		token, err := oauthSetup.OAuth2.Exchange(ctx, body.Code,
			oauth2.SetAuthURLParam("redirect_uri", strings.TrimSpace(body.RedirectURI)),
			oauth2.SetAuthURLParam("code_verifier", body.CodeVerifier),
		)
		if err != nil {
			logger.Warn("discord token exchange failed", zap.Error(err))
			return echo.NewHTTPError(400, "Could not complete login with Discord. Please try again.")
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
		if resp.StatusCode != 200 {
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

		encAccess, err := oauthSetup.TokenEncryptor.Encrypt(token.AccessToken)
		if err != nil {
			logger.Error("encrypt discord access token", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot issue session.")
		}
		encDiscordRefresh, err := oauthSetup.TokenEncryptor.Encrypt(token.RefreshToken)
		if err != nil {
			logger.Error("encrypt discord refresh token", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot issue session.")
		}

		if err := userSessionRepo.WithContext(ctx).DeleteByDiscordUserID(ctx, du.ID); err != nil {
			logger.Error("delete prior user sessions", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot issue session.")
		}
		logger.Debug("received user info", zap.Any("userInfo", du))
		sess := &models.UserSession{
			DiscordUserID:                du.ID,
			RefreshTokenHash:             refreshHash,
			RefreshTokenExpiry:           expiry,
			EncryptedDiscordAccessToken:  encAccess,
			EncryptedDiscordRefreshToken: encDiscordRefresh,
			DiscordTokenExpiry:           token.Expiry,
		}
		if err := userSessionRepo.WithContext(ctx).CreateSession(ctx, sess); err != nil {
			logger.Error("save user session", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot issue session.")
		}

		return c.JSON(200, tokenResponse{
			AccessToken:  accessJWT,
			RefreshToken: refreshPlain,
			ExpiresIn:    int64(auth.DefaultAccessTokenTTL.Seconds()),
		})
	})

	e.POST("/auth/logout", func(c echo.Context) error {
		ctx := c.Request().Context()
		authContext := c.(*apimw.AuthContext)
		if err := userSessionRepo.WithContext(ctx).DeleteByDiscordUserID(ctx, authContext.UserID); err != nil {
			logger.Error("delete user session on logout", zap.Error(err))
			return echo.NewHTTPError(500, "Cannot log out.")
		}
		return c.NoContent(204)
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
		return c.JSON(200, refreshTokenResponse{
			AccessToken:  accessJWT,
			RefreshToken: newPlain,
			ExpiresIn:    int64(auth.DefaultAccessTokenTTL.Seconds()),
		})
	})
}
