package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type oauthPending struct {
	codeVerifier string
	expiresAt    time.Time
}

var oauthStateMu sync.Mutex
var oauthStateStore = make(map[string]*oauthPending)

type authHandlerDeps struct {
	db              *repositories.UserSessionRepositoryImpl
	dg              *discordgo.Session
	logger          *zap.Logger
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type refreshRequestBody struct {
	RefreshToken string `json:"refresh_token"`
}

type guildsResponse struct {
	Guilds []guildJSON `json:"guilds"`
}

type guildJSON struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Icon *string `json:"icon"`
}

func registerOAuthAndGuildRoutes(
	e *echo.Echo,
	db *repositories.UserSessionRepositoryImpl,
	dg *discordgo.Session,
	logger *zap.Logger,
) {
	if !oauthEnvComplete() {
		logger.Info("OAuth2 routes (/auth/*, /guilds) not registered: missing DISCORD_* or JWT_SIGNING_KEY.")
		return
	}
	deps := &authHandlerDeps{db: db, dg: dg, logger: logger.Named("api.oauth")}

	e.GET("/auth/discord", func(c echo.Context) error {
		verifier, err := randomPKCEVerifier()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not start OAuth.")
		}
		stateKey, err := randomOAuthStateKey()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not start OAuth.")
		}
		oauthStateMu.Lock()
		for k, v := range oauthStateStore {
			if time.Now().After(v.expiresAt) {
				delete(oauthStateStore, k)
			}
		}
		oauthStateStore[stateKey] = &oauthPending{
			codeVerifier: verifier,
			expiresAt:    time.Now().Add(10 * time.Minute),
		}
		oauthStateMu.Unlock()

		cfg := discordOAuthConfig()
		url := cfg.AuthCodeURL(stateKey,
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("code_challenge", pkceChallengeS256(verifier)),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
		return c.Redirect(http.StatusFound, url)
	})

	e.GET("/auth/discord/callback", func(c echo.Context) error {
		state := c.QueryParam("state")
		code := c.QueryParam("code")
		if state == "" || code == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Missing state or code.")
		}
		oauthStateMu.Lock()
		pending, ok := oauthStateStore[state]
		if ok {
			delete(oauthStateStore, state)
		}
		oauthStateMu.Unlock()
		if !ok || pending == nil || time.Now().After(pending.expiresAt) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired state.")
		}

		cfg := discordOAuthConfig()
		ctx := c.Request().Context()
		tok, err := cfg.Exchange(ctx, code,
			oauth2.SetAuthURLParam("code_verifier", pending.codeVerifier),
		)
		if err != nil {
			deps.logger.Warn("Discord code exchange failed", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadGateway, "Discord token exchange failed.")
		}

		u, newTok, err := fetchDiscordUserMe(ctx, tok.AccessToken, tok.RefreshToken, &tok.Expiry)
		if err != nil {
			deps.logger.Warn("Discord @me failed after exchange", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadGateway, "Could not load Discord user.")
		}
		tok = newTok

		accessEnc, err := encryptToken(tok.AccessToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Session storage error.")
		}
		refreshEnc, err := encryptToken(tok.RefreshToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Session storage error.")
		}

		groRefresh, groHash, err := newGroRefreshCredentials()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not issue refresh token.")
		}
		now := time.Now()
		expiresGro := now.Add(groRefreshTTL())
		var discExp *time.Time
		if !tok.Expiry.IsZero() {
			t := tok.Expiry
			discExp = &t
		}

		sess := &models.UserSession{
			DiscordUserID:          u.ID,
			DiscordAccessTokenEnc:  accessEnc,
			DiscordRefreshTokenEnc: refreshEnc,
			DiscordTokenExpiresAt:  discExp,
			GroRefreshTokenHash:    groHash,
			GroRefreshExpiresAt:    expiresGro,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := deps.db.WithContext(ctx).Upsert(ctx, sess); err != nil {
			return err
		}

		accessJWT, err := signAccessJWT(u.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not sign token.")
		}

		return c.JSON(http.StatusOK, tokenResponse{
			AccessToken:  accessJWT,
			RefreshToken: groRefresh,
			TokenType:    "Bearer",
			ExpiresIn:    int64(jwtAccessTTL().Seconds()),
		})
	})

	e.POST("/auth/refresh", func(c echo.Context) error {
		var body refreshRequestBody
		if err := c.Bind(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON body.")
		}
		if strings.TrimSpace(body.RefreshToken) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "refresh_token is required.")
		}
		ctx := c.Request().Context()
		hash := hashGroRefreshToken(body.RefreshToken)
		row, err := deps.db.WithContext(ctx).GetByGroRefreshTokenHash(ctx, hash)
		if err != nil {
			return err
		}
		if row == nil || time.Now().After(row.GroRefreshExpiresAt) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired refresh token.")
		}

		accessJWT, err := signAccessJWT(row.DiscordUserID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not sign token.")
		}

		return c.JSON(http.StatusOK, tokenResponse{
			AccessToken:  accessJWT,
			RefreshToken: body.RefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(jwtAccessTTL().Seconds()),
		})
	})

	e.GET("/guilds", func(c echo.Context) error {
		authCtx, ok := c.(*AuthContext)
		if !ok || authCtx.AuthScheme != "Bearer" || authCtx.UserID == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Bearer authentication required.")
		}
		ctx := c.Request().Context()
		row, err := deps.db.WithContext(ctx).GetByDiscordUserID(ctx, authCtx.UserID)
		if err != nil {
			return err
		}
		if row == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Session not found; please sign in again.")
		}

		access, err := decryptToken(row.DiscordAccessTokenEnc)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Session decrypt error.")
		}
		refresh, err := decryptToken(row.DiscordRefreshTokenEnc)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Session decrypt error.")
		}

		guilds, newTok, err := fetchDiscordMyGuilds(ctx, access, refresh, row.DiscordTokenExpiresAt)
		if err != nil {
			deps.logger.Warn("Discord guilds fetch failed", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, "Discord session expired; please sign in again.")
		}

		if newTok != nil && newTok.AccessToken != access {
			accessEnc, err := encryptToken(newTok.AccessToken)
			if err != nil {
				return err
			}
			refreshEnc, err := encryptToken(newTok.RefreshToken)
			if err != nil {
				return err
			}
			var discExp *time.Time
			if !newTok.Expiry.IsZero() {
				t := newTok.Expiry
				discExp = &t
			}
			row.DiscordAccessTokenEnc = accessEnc
			row.DiscordRefreshTokenEnc = refreshEnc
			row.DiscordTokenExpiresAt = discExp
			row.UpdatedAt = time.Now()
			if err := deps.db.WithContext(ctx).Upsert(ctx, row); err != nil {
				return err
			}
		}

		botGuilds := botInstalledGuildIDs(deps.dg)
		out := make([]guildJSON, 0)
		for _, g := range guilds {
			if botGuilds[g.ID] {
				out = append(out, guildJSON{ID: g.ID, Name: g.Name, Icon: g.Icon})
			}
		}
		return c.JSON(http.StatusOK, guildsResponse{Guilds: out})
	})
}

func botInstalledGuildIDs(s *discordgo.Session) map[string]bool {
	m := make(map[string]bool)
	if s == nil {
		return m
	}
	for _, g := range s.State.Guilds {
		m[g.ID] = true
	}
	return m
}

func newGroRefreshCredentials() (plaintext string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	plaintext = hex.EncodeToString(b)
	hash = hashGroRefreshToken(plaintext)
	return plaintext, hash, nil
}

func hashGroRefreshToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

func jwtHTTPError(err error) *echo.HTTPError {
	if errors.Is(err, jwt.ErrTokenExpired) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Token expired.")
	}
	if errors.Is(err, jwt.ErrTokenMalformed) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Malformed token.")
	}
	if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		return echo.NewHTTPError(http.StatusForbidden, "Invalid token signature.")
	}
	return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token.")
}
