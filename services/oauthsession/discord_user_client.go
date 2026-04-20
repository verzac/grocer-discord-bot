package oauthsession

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// DiscordUserHTTPClient returns an HTTP client that sends requests with the user's Discord OAuth access token.
// It refreshes the token via oauth2.TokenSource when needed and best-effort persists rotated tokens to user_sessions.
func (s *impl) DiscordUserHTTPClient(ctx context.Context, discordUserID string) (*http.Client, error) {
	sess, err := s.repo.WithContext(ctx).FindByDiscordUserID(ctx, discordUserID)
	if err != nil {
		s.log.Error("load user session for discord oauth", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Cannot load session.")
	}
	if sess == nil || sess.EncryptedDiscordAccessToken == "" {
		return nil, ErrSessionNotFound
	}

	accessPlain, err := s.oauth.TokenEncryptor.Decrypt(sess.EncryptedDiscordAccessToken)
	if err != nil {
		s.log.Error("decrypt discord access token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Cannot load session.")
	}
	var refreshPlain string
	if sess.EncryptedDiscordRefreshToken != "" {
		refreshPlain, err = s.oauth.TokenEncryptor.Decrypt(sess.EncryptedDiscordRefreshToken)
		if err != nil {
			s.log.Error("decrypt discord refresh token", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "Cannot load session.")
		}
	}

	stored := &oauth2.Token{
		AccessToken:  accessPlain,
		RefreshToken: refreshPlain,
		Expiry:       sess.DiscordTokenExpiry,
	}
	ts := s.oauth.OAuth2.TokenSource(ctx, stored)
	fresh, err := ts.Token()
	if err != nil {
		s.log.Warn("discord token refresh or validate failed", zap.Error(err))
		return nil, ErrDiscordTokenInvalid
	}

	s.persistRotatedTokens(ctx, sess, accessPlain, refreshPlain, fresh)

	return s.oauth.OAuth2.Client(ctx, fresh), nil
}

func (s *impl) persistRotatedTokens(ctx context.Context, sess *models.UserSession, accessPlain, refreshPlain string, fresh *oauth2.Token) {
	if fresh.AccessToken == accessPlain && fresh.RefreshToken == refreshPlain {
		return
	}
	encAccess, encErr := s.oauth.TokenEncryptor.Encrypt(fresh.AccessToken)
	if encErr != nil {
		s.log.Error("encrypt rotated discord access token", zap.Error(encErr))
		return
	}
	sess.EncryptedDiscordAccessToken = encAccess
	if fresh.RefreshToken != "" {
		encRef, refErr := s.oauth.TokenEncryptor.Encrypt(fresh.RefreshToken)
		if refErr != nil {
			s.log.Error("encrypt rotated discord refresh token", zap.Error(refErr))
			return
		}
		sess.EncryptedDiscordRefreshToken = encRef
		if !fresh.Expiry.IsZero() {
			sess.DiscordTokenExpiry = fresh.Expiry
		}
		if upErr := s.repo.WithContext(ctx).UpdateSession(ctx, sess); upErr != nil {
			s.log.Warn("persist rotated discord tokens", zap.Error(upErr))
		}
		return
	}
	if !fresh.Expiry.IsZero() {
		sess.DiscordTokenExpiry = fresh.Expiry
	}
	if upErr := s.repo.WithContext(ctx).UpdateSession(ctx, sess); upErr != nil {
		s.log.Warn("persist rotated discord tokens", zap.Error(upErr))
	}
}
