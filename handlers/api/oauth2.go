package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	cache "github.com/patrickmn/go-cache"
)

const (
	oauthStateTTL        = 5 * time.Minute
	refreshTokenTTL      = 7 * 24 * time.Hour
	oauthStateCacheKeyPF = "oauth_state:"
)

var oauthStateCache = cache.New(oauthStateTTL, 10*time.Minute)

type oauthDeps struct {
	logger          *zap.Logger
	userSessionRepo repositories.UserSessionRepository
	oauth2Cfg       *oauth2.Config
	jwtKey          string
	encryptionKey   string
}

func newOAuthDeps(logger *zap.Logger, userSessionRepo repositories.UserSessionRepository, cfg *oauth2.Config, jwtKey string) *oauthDeps {
	return &oauthDeps{
		logger:          logger.Named("oauth2"),
		userSessionRepo: userSessionRepo,
		oauth2Cfg:       cfg,
		jwtKey:          jwtKey,
		encryptionKey:   jwtKey,
	}
}

func discordOAuth2ConfigFromEnv() *oauth2.Config {
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	secret := os.Getenv("DISCORD_CLIENT_SECRET")
	redirect := os.Getenv("DISCORD_REDIRECT_URI")
	if clientID == "" || secret == "" || redirect == "" {
		return nil
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		RedirectURL:  redirect,
		Scopes:       []string{"identify", "guilds"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
}

func oauthEnvComplete() bool {
	return os.Getenv("DISCORD_CLIENT_ID") != "" &&
		os.Getenv("DISCORD_CLIENT_SECRET") != "" &&
		os.Getenv("DISCORD_REDIRECT_URI") != ""
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randomCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func pkceChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (d *oauthDeps) handleAuthDiscordStart(c echo.Context) error {
	state, err := randomHex(16)
	if err != nil {
		return err
	}
	verifier, err := randomCodeVerifier()
	if err != nil {
		return err
	}
	challenge := pkceChallengeS256(verifier)
	oauthStateCache.Set(oauthStateCacheKeyPF+state, verifier, oauthStateTTL)
	url := d.oauth2Cfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	return c.Redirect(http.StatusFound, url)
}

func (d *oauthDeps) handleAuthDiscordCallback(c echo.Context) error {
	ctx := c.Request().Context()
	state := c.QueryParam("state")
	code := c.QueryParam("code")
	if state == "" || code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing state or code.")
	}
	raw, ok := oauthStateCache.Get(oauthStateCacheKeyPF + state)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired state.")
	}
	oauthStateCache.Delete(oauthStateCacheKeyPF + state)
	codeVerifier, ok := raw.(string)
	if !ok || codeVerifier == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid state payload.")
	}
	tok, err := d.oauth2Cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		d.logger.Warn("Discord token exchange failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadGateway, "Discord token exchange failed.")
	}
	user, err := fetchDiscordUserMe(ctx, tok.AccessToken)
	if err != nil {
		d.logger.Error("Discord users/@me failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadGateway, "Failed to load Discord profile.")
	}
	encAccess, err := auth.EncryptWithKey(d.encryptionKey, tok.AccessToken)
	if err != nil {
		return err
	}
	encRefresh, err := auth.EncryptWithKey(d.encryptionKey, tok.RefreshToken)
	if err != nil {
		return err
	}
	refreshPlain, refreshHash, err := randomRefreshPair()
	if err != nil {
		return err
	}
	now := time.Now()
	session := &models.UserSession{
		DiscordUserID:       user.ID,
		DiscordAccessToken:  encAccess,
		DiscordRefreshToken: encRefresh,
		DiscordTokenExpiry:  discordTokenExpiry(tok),
		RefreshTokenHash:    refreshHash,
		RefreshTokenExpiry:  now.Add(refreshTokenTTL),
	}
	if err := d.userSessionRepo.UpsertByDiscordUserID(ctx, session); err != nil {
		return err
	}
	accessJWT, err := issueAccessToken(d.jwtKey, user.ID, defaultAccessTokenTTL)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token":  accessJWT,
		"refresh_token": refreshPlain,
		"token_type":    "Bearer",
		"expires_in":    int(defaultAccessTokenTTL.Seconds()),
	})
}

func randomRefreshPair() (plaintext string, sha256Hex string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	plain := hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plain))
	return plain, hex.EncodeToString(sum[:]), nil
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (d *oauthDeps) handleAuthRefresh(c echo.Context) error {
	ctx := c.Request().Context()
	var body refreshRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil || body.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing refresh_token.")
	}
	sum := sha256.Sum256([]byte(body.RefreshToken))
	hash := hex.EncodeToString(sum[:])
	row, err := d.userSessionRepo.GetByRefreshTokenHash(ctx, hash)
	if err != nil {
		return err
	}
	if row == nil || time.Now().After(row.RefreshTokenExpiry) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired refresh token.")
	}
	accessJWT, err := issueAccessToken(d.jwtKey, row.DiscordUserID, defaultAccessTokenTTL)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": accessJWT,
		"token_type":   "Bearer",
		"expires_in":   int(defaultAccessTokenTTL.Seconds()),
	})
}
