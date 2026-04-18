package auth

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var discordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

// OAuthSetup holds Discord OAuth2 client config and redirect URI allowlist for token exchange.
type OAuthSetup struct {
	OAuth2              *oauth2.Config
	AllowedRedirectURIs []string
	TokenEncryptor      *TokenEncryptor
}

// ValidateRedirectURI returns true if uri exactly matches one of the allowed redirect URIs (after trim).
func (s *OAuthSetup) ValidateRedirectURI(uri string) bool {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return false
	}
	for _, allowed := range s.AllowedRedirectURIs {
		if allowed == uri {
			return true
		}
	}
	return false
}

func parseAllowedRedirectURIsFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("ALLOWED_REDIRECT_URIS"))
	if raw == "" {
		return []string{"grocerybot://auth/callback"}
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"grocerybot://auth/callback"}
	}
	return out
}

// LoadOAuthSetup reads Discord OAuth env vars and builds an oauth2.Config.
// Returns nil if any required Discord env var is missing (caller should skip /auth routes).
func LoadOAuthSetup(logger *zap.Logger) *OAuthSetup {
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	secret := os.Getenv("DISCORD_CLIENT_SECRET")
	if clientID == "" || secret == "" {
		logger.Warn("Discord OAuth2 env vars incomplete; /auth routes will not be registered",
			zap.Bool("has_client_id", clientID != ""),
			zap.Bool("has_client_secret", secret != ""),
		)
		return nil
	}
	sessionKey := os.Getenv("SESSION_ENCRYPTION_KEY")
	encryptor, err := NewTokenEncryptor(sessionKey)
	if err != nil {
		logger.Warn("SESSION_ENCRYPTION_KEY missing or invalid; /auth routes will not be registered",
			zap.Error(err),
		)
		return nil
	}

	allowed := parseAllowedRedirectURIsFromEnv()
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		RedirectURL:  "",
		Scopes:       []string{"identify", "guilds"},
		Endpoint:     discordEndpoint,
	}
	return &OAuthSetup{
		OAuth2:              cfg,
		AllowedRedirectURIs: allowed,
		TokenEncryptor:      encryptor,
	}
}
