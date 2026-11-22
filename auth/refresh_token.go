package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

const DefaultRefreshTokenTTL = 7 * 24 * time.Hour

// GenerateRefreshToken returns a URL-safe plaintext refresh token for the client and a hex-encoded SHA-256 hash for storage.
func GenerateRefreshToken() (plaintext string, hashHex string, err error) {
	plaintext, err = randomURLToken()
	if err != nil {
		return "", "", fmt.Errorf("refresh token: %w", err)
	}
	return plaintext, HashRefreshToken(plaintext), nil
}

// HashRefreshToken returns a hex-encoded SHA-256 of the plaintext refresh token.
func HashRefreshToken(token string) string {
	return sha256Hex(token)
}

// randomURLToken returns 32 random bytes encoded as unpadded base64url.
func randomURLToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
