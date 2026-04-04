package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultAccessTokenTTL = 15 * time.Minute

func issueAccessToken(signingKey string, discordUserID string, ttl time.Duration) (string, error) {
	if signingKey == "" {
		return "", errors.New("missing signing key")
	}
	if ttl <= 0 {
		ttl = defaultAccessTokenTTL
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   discordUserID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(signingKey))
}

func parseAccessToken(signingKey string, tokenString string) (discordUserID string, err error) {
	if signingKey == "" {
		return "", errors.New("missing signing key")
	}
	tok, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(signingKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", errJWTExpired
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return "", errJWTSignatureInvalid
		}
		return "", errJWTMalformed
	}
	claims, ok := tok.Claims.(*jwt.RegisteredClaims)
	if !ok || !tok.Valid {
		return "", errJWTMalformed
	}
	if claims.Subject == "" {
		return "", errJWTMalformed
	}
	return claims.Subject, nil
}

var (
	errJWTExpired          = errors.New("jwt expired")
	errJWTSignatureInvalid = errors.New("jwt signature invalid")
	errJWTMalformed        = errors.New("jwt malformed")
)

func IsJWTExpired(err error) bool {
	return errors.Is(err, errJWTExpired)
}

func IsJWTSignatureInvalid(err error) bool {
	return errors.Is(err, errJWTSignatureInvalid)
}

func IsJWTMalformed(err error) bool {
	return errors.Is(err, errJWTMalformed)
}
