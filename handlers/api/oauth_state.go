package api

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func randomOAuthStateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
