package auth

import (
	"fmt"
	"time"
)

const (
	// DefaultIntentTokenTTL is a fixed lifetime; intent tokens are not extended on use.
	DefaultIntentTokenTTL = 30 * 24 * time.Hour
	// IntentTokenPrefix distinguishes intent tokens from JWTs (which start with "eyJ") in the Bearer header.
	IntentTokenPrefix = "gbi_"
)

// GenerateIntentToken returns a plaintext intent token ("gbi_" + base64url of 32 random bytes) for the client
// and a hex-encoded SHA-256 hash of the full plaintext for storage.
func GenerateIntentToken() (plaintext string, hashHex string, err error) {
	raw, err := randomURLToken()
	if err != nil {
		return "", "", fmt.Errorf("intent token: %w", err)
	}
	plaintext = IntentTokenPrefix + raw
	return plaintext, HashIntentToken(plaintext), nil
}

// HashIntentToken returns a hex-encoded SHA-256 of the full plaintext intent token, prefix included.
func HashIntentToken(token string) string {
	return sha256Hex(token)
}
