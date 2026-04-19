package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const gcmNonceSize = 12

// TokenEncryptor encrypts and decrypts strings with AES-GCM (random nonce per encrypt).
type TokenEncryptor struct {
	key []byte
}

// NewTokenEncryptor builds an encryptor from a hex-encoded 32-byte AES key.
func NewTokenEncryptor(keyHex string) (*TokenEncryptor, error) {
	if keyHex == "" {
		return nil, errors.New("session encryption key is empty")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("session encryption key: invalid hex: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("session encryption key must decode to 32 bytes, got %d", len(key))
	}
	return &TokenEncryptor{key: key}, nil
}

// Encrypt returns base64.StdEncoding of nonce || ciphertext.
func (e *TokenEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcmNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt reverses Encrypt.
func (e *TokenEncryptor) Decrypt(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("ciphertext: decode: %w", err)
	}
	if len(raw) < gcmNonceSize {
		return "", errors.New("ciphertext too short")
	}
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := raw[:gcmNonceSize]
	ciphertext := raw[gcmNonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
