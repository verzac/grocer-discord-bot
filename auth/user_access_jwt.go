package auth

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const DefaultAccessTokenTTL = 15 * time.Minute

type UserInfoJWT struct {
	DiscordUserID string
}

type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
}

var (
	DefaultJWTIssuer *JWTIssuer
	hasInit          = sync.Once{}
)

func NewJWTIssuer(secret []byte) (*JWTIssuer, error) {
	ttl := DefaultAccessTokenTTL
	if len(secret) == 0 {
		return nil, errors.New("JWT signing key is empty")
	}
	if ttl <= 0 {
		return nil, errors.New("TTL must be positive")
	}
	return &JWTIssuer{secret: secret, ttl: ttl}, nil
}

func InitDefaultJWTIssuer(logger *zap.Logger) {
	hasInit.Do(func() {
		secret := os.Getenv("JWT_SIGNING_KEY")
		if secret == "" {
			logger.Error("JWT_SIGNING_KEY is not set")
		}
		jwtIssuer, err := NewJWTIssuer([]byte(secret))
		if err != nil {
			logger.Error("cannot create default JWT issuer", zap.Error(err))
		}
		DefaultJWTIssuer = jwtIssuer
	})
}

type UserJWTClaims struct {
	jwt.RegisteredClaims
}

func (u *JWTIssuer) Issue(ctx context.Context, discordUserID string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if discordUserID == "" {
		return "", errors.New("discord user id is empty")
	}
	now := time.Now()
	claims := UserJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   discordUserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(u.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(u.secret)
}

func (u *JWTIssuer) Verify(ctx context.Context, tokenString string) (*UserInfoJWT, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if tokenString == "" {
		return nil, errors.New("where is token?")
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	var claims UserJWTClaims
	_, err := parser.ParseWithClaims(tokenString, &claims, func(*jwt.Token) (interface{}, error) {
		return u.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims.Subject == "" {
		return nil, errors.New("token is missing required subject claim")
	}
	return &UserInfoJWT{DiscordUserID: claims.Subject}, nil
}

func VerifyErrorHTTPStatus(err error) int {
	if err == nil {
		return 200
	}
	if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		return 403
	}
	return 401
}
