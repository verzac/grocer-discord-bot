package demo

import (
	"context"
	"crypto/subtle"
	"os"

	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
)

type LoginResult struct {
	AccessToken string
	ExpiresIn   int64
}

var Service *DemoService

type DemoService struct {
	demoUserID  string
	demoPassword string
	userSessionRepo repositories.UserSessionRepository
	logger          *zap.Logger
}

func Init(logger *zap.Logger, userSessionRepo repositories.UserSessionRepository) {
	demoUserID := os.Getenv("DEMO_USER_DISCORD_ID")
	demoPassword := os.Getenv("DEMO_LOGIN_PASSWORD")
	if demoUserID == "" || demoPassword == "" {
		return
	}
	Service = &DemoService{
		demoUserID:      demoUserID,
		demoPassword:    demoPassword,
		userSessionRepo: userSessionRepo,
		logger:          logger.Named("demo"),
	}
	logger.Info("demo mode enabled", zap.String("demoUserID", demoUserID))
}

func (s *DemoService) IsDemoUser(userID string) bool {
	return s.demoUserID == userID
}

func (s *DemoService) Login(ctx context.Context, password string) (*LoginResult, error) {
	if subtle.ConstantTimeCompare([]byte(password), []byte(s.demoPassword)) != 1 {
		return nil, ErrInvalidPassword
	}

	sess, err := s.userSessionRepo.WithContext(ctx).FindByDiscordUserID(ctx, s.demoUserID)
	if err != nil {
		s.logger.Error("lookup demo session", zap.Error(err))
		return nil, ErrInternal
	}
	if sess == nil {
		s.logger.Error("demo account session not found - login with the demo Discord account first")
		return nil, ErrNotSeeded
	}

	if auth.DefaultJWTIssuer == nil {
		return nil, ErrInternal
	}
	accessJWT, err := auth.DefaultJWTIssuer.Issue(ctx, s.demoUserID)
	if err != nil {
		s.logger.Error("issue access jwt", zap.Error(err))
		return nil, ErrInternal
	}

	return &LoginResult{
		AccessToken: accessJWT,
		ExpiresIn:   int64(auth.DefaultAccessTokenTTL.Seconds()),
	}, nil
}
