package oauthsession

import (
	"context"
	"net/http"

	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
)

var Service OAuthSessionService

type OAuthSessionService interface {
	DiscordUserHTTPClient(ctx context.Context, discordUserID string) (*http.Client, error)
}

type impl struct {
	oauth *auth.OAuthSetup
	repo  repositories.UserSessionRepository
	log   *zap.Logger
}

// Init wires the global Service used by handlers (e.g. GET /guilds). Call when OAuth env is complete.
func Init(oauthSetup *auth.OAuthSetup, userSessionRepo repositories.UserSessionRepository, logger *zap.Logger) {
	Service = &impl{
		oauth: oauthSetup,
		repo:  userSessionRepo,
		log:   logger.Named("oauthsession"),
	}
}
