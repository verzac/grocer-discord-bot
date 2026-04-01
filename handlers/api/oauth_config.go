package api

import (
	"os"

	"golang.org/x/oauth2"
)

func oauthEnvComplete() bool {
	return os.Getenv("DISCORD_CLIENT_ID") != "" &&
		os.Getenv("DISCORD_CLIENT_SECRET") != "" &&
		os.Getenv("DISCORD_REDIRECT_URI") != "" &&
		os.Getenv("JWT_SIGNING_KEY") != ""
}

func discordOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("DISCORD_REDIRECT_URI"),
		Scopes:       []string{"identify", "guilds"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
}
