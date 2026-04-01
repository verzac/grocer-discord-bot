package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"os"
)

func sessionAEAD() (cipher.AEAD, error) {
	keyStr := os.Getenv("JWT_SIGNING_KEY")
	if keyStr == "" {
		return nil, errors.New("JWT_SIGNING_KEY is not set")
	}
	sum := sha256.Sum256([]byte(keyStr))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func encryptToken(plaintext string) ([]byte, error) {
	aead, err := sessionAEAD()
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return aead.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func decryptToken(ciphertext []byte) (string, error) {
	aead, err := sessionAEAD()
	if err != nil {
		return "", err
	}
	ns := aead.NonceSize()
	if len(ciphertext) < ns {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := ciphertext[:ns], ciphertext[ns:]
	pt, err := aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}
