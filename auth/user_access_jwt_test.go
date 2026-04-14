package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestUserAccessJWT_roundTrip(t *testing.T) {
	secret := []byte("unit-test-signing-key-32bytes!!")
	issuer, err := NewJWTIssuer(secret)
	require.NoError(t, err)
	ctx := context.Background()

	token, err := issuer.Issue(ctx, "183947835467759617")
	require.NoError(t, err)

	got, err := issuer.Verify(ctx, token)
	require.NoError(t, err)
	require.Equal(t, "183947835467759617", got.DiscordUserID)
}

func TestUserAccessJWT_wrongSecret_returns403Mapped(t *testing.T) {
	secretA := []byte("unit-test-signing-key-32bytes-a!")
	secretB := []byte("unit-test-signing-key-32bytes-b!")
	issuerA, err := NewJWTIssuer(secretA)
	require.NoError(t, err)
	issuerB, err := NewJWTIssuer(secretB)
	require.NoError(t, err)
	ctx := context.Background()

	token, err := issuerA.Issue(ctx, "sub123")
	require.NoError(t, err)

	_, err = issuerB.Verify(ctx, token)
	require.Error(t, err)
	require.Equal(t, 403, VerifyErrorHTTPStatus(err))
	require.True(t, errors.Is(err, jwt.ErrTokenSignatureInvalid))
}

func TestUserAccessJWT_bogusString_returns401(t *testing.T) {
	secret := []byte("unit-test-signing-key-32bytes!!")
	issuer, err := NewJWTIssuer(secret)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = issuer.Verify(ctx, "not-a-jwt")
	require.Error(t, err)
	require.Equal(t, 401, VerifyErrorHTTPStatus(err))
}

func TestUserAccessJWT_expired_returns401(t *testing.T) {
	secret := []byte("unit-test-signing-key-32bytes!!")
	issuer, err := NewJWTIssuer(secret)
	require.NoError(t, err)
	ctx := context.Background()

	past := time.Unix(100, 0)
	claims := UserJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "sub123",
			IssuedAt:  jwt.NewNumericDate(past),
			ExpiresAt: jwt.NewNumericDate(past.Add(time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	require.NoError(t, err)

	_, err = issuer.Verify(ctx, signed)
	require.Error(t, err)
	require.Equal(t, 401, VerifyErrorHTTPStatus(err))
	require.True(t, errors.Is(err, jwt.ErrTokenExpired))
}
