package api

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultJWTAccessTTLMin   = 15
	defaultGroRefreshTTLDays = 7
)

func jwtSigningKey() ([]byte, error) {
	k := os.Getenv("JWT_SIGNING_KEY")
	if k == "" {
		return nil, errors.New("JWT_SIGNING_KEY is not set")
	}
	return []byte(k), nil
}

func jwtAccessTTL() time.Duration {
	if v := os.Getenv("GROCER_JWT_ACCESS_TTL_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Minute
		}
	}
	return defaultJWTAccessTTLMin * time.Minute
}

func groRefreshTTL() time.Duration {
	if v := os.Getenv("GROCER_GRO_REFRESH_TTL_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * 24 * time.Hour
		}
	}
	return defaultGroRefreshTTLDays * 24 * time.Hour
}

func signAccessJWT(discordUserID string) (string, error) {
	key, err := jwtSigningKey()
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   discordUserID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(jwtAccessTTL())),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(key)
}

func parseAccessJWT(tokenStr string) (*jwt.RegisteredClaims, error) {
	key, err := jwtSigningKey()
	if err != nil {
		return nil, err
	}
	t, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*jwt.RegisteredClaims)
	if !ok || !t.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
