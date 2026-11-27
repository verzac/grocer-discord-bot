package auth

import (
	"fmt"
	"os"
	"time"
)

const (
	// DefaultIntentTokenTTL is a fixed lifetime; intent tokens are not extended on use.
	DefaultIntentTokenTTL = 30 * 24 * time.Hour
	// IntentTokenPrefix distinguishes intent tokens from JWTs (which start with "eyJ") in the Bearer header.
	IntentTokenPrefix = "gbi_"
)

// IntentTokensDisabled reports whether the DISABLE_INTENT_TOKENS kill switch is set to "true". When set, minting is
// refused and existing intent tokens are rejected; stored rows are kept. The process environment is fixed at startup,
// so changing it requires a restart.
func IntentTokensDisabled() bool {
	return os.Getenv("DISABLE_INTENT_TOKENS") == "true"
}

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
