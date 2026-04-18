package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenEncryptor_roundTrip(t *testing.T) {
	keyHex := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	e, err := NewTokenEncryptor(keyHex)
	require.NoError(t, err)

	plain := "discord-access-token-secret"
	enc, err := e.Encrypt(plain)
	require.NoError(t, err)
	require.NotEqual(t, plain, enc)

	got, err := e.Decrypt(enc)
	require.NoError(t, err)
	require.Equal(t, plain, got)
}

func TestTokenEncryptor_wrongKey_failsDecrypt(t *testing.T) {
	keyA := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	keyB := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	ea, err := NewTokenEncryptor(keyA)
	require.NoError(t, err)
	eb, err := NewTokenEncryptor(keyB)
	require.NoError(t, err)

	enc, err := ea.Encrypt("hello")
	require.NoError(t, err)
	_, err = eb.Decrypt(enc)
	require.Error(t, err)
}

func TestTokenEncryptor_tampered_failsDecrypt(t *testing.T) {
	keyHex := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	e, err := NewTokenEncryptor(keyHex)
	require.NoError(t, err)

	enc, err := e.Encrypt("intact")
	require.NoError(t, err)
	if len(enc) > 4 {
		enc = enc[:len(enc)-4] + "AAAA"
	}
	_, err = e.Decrypt(enc)
	require.Error(t, err)
}

func TestTokenEncryptor_emptyKey_errors(t *testing.T) {
	_, err := NewTokenEncryptor("")
	require.Error(t, err)
}
