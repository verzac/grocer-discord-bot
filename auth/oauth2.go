package auth

import (
	"os"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var discordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

// OAuthSetup holds Discord OAuth2 client config and the post-login redirect for the app.
type OAuthSetup struct {
	OAuth2         *oauth2.Config
	AppRedirectURI string
}

// LoadOAuthSetup reads Discord OAuth env vars and builds an oauth2.Config.
// Returns nil if any required Discord env var is missing (caller should skip /auth routes).
func LoadOAuthSetup(logger *zap.Logger) *OAuthSetup {
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	secret := os.Getenv("DISCORD_CLIENT_SECRET")
	redirect := os.Getenv("DISCORD_REDIRECT_URI")
	if clientID == "" || secret == "" || redirect == "" {
		logger.Warn("Discord OAuth2 env vars incomplete; /auth routes will not be registered",
			zap.Bool("has_client_id", clientID != ""),
			zap.Bool("has_client_secret", secret != ""),
			zap.Bool("has_redirect_uri", redirect != ""),
		)
		return nil
	}
	appRedirect := os.Getenv("APP_REDIRECT_URI")
	if appRedirect == "" {
		appRedirect = "grocerybot://auth/callback"
	}
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		RedirectURL:  redirect,
		Scopes:       []string{"identify", "guilds"},
		Endpoint:     discordEndpoint,
	}
	return &OAuthSetup{
		OAuth2:         cfg,
		AppRedirectURI: appRedirect,
	}
}
