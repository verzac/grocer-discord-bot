package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
)

type fakeIntentTokenRepo struct {
	byHash  map[string]*models.IntentToken
	lookups int
}

var _ repositories.IntentTokenRepository = &fakeIntentTokenRepo{}

func (f *fakeIntentTokenRepo) WithContext(ctx context.Context) repositories.IntentTokenRepository {
	return f
}

func (f *fakeIntentTokenRepo) Create(ctx context.Context, token *models.IntentToken) error {
	return errors.New("not implemented")
}

func (f *fakeIntentTokenRepo) FindByTokenHash(ctx context.Context, hash string) (*models.IntentToken, error) {
	f.lookups++
	return f.byHash[hash], nil
}

func (f *fakeIntentTokenRepo) DeleteByDiscordUserID(ctx context.Context, discordUserID string) error {
	return errors.New("not implemented")
}

const (
	validIntentToken   = "gbi_valid"
	expiredIntentToken = "gbi_expired"
)

func newFakeIntentTokenRepo() *fakeIntentTokenRepo {
	return &fakeIntentTokenRepo{byHash: map[string]*models.IntentToken{
		auth.HashIntentToken(validIntentToken): {
			ID: 1, DiscordUserID: "user-1", ExpiresAt: time.Now().UTC().Add(time.Hour),
		},
		auth.HashIntentToken(expiredIntentToken): {
			ID: 2, DiscordUserID: "user-2", ExpiresAt: time.Now().UTC().Add(-time.Second),
		},
	}}
}

// runAuth runs AuthMiddleware for one request and reports the HTTP status the middleware chose
// (200 if it called next).
func runAuth(t *testing.T, repo *fakeIntentTokenRepo, method, path, authHeader string) (status int, nextCalled bool) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := AuthMiddleware(nil, repo, zap.NewNop(), "test", nil)(func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	})
	err := h(c)
	if err == nil {
		return rec.Code, nextCalled
	}
	var he *echo.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("unexpected non-HTTP error: %v", err)
	}
	return he.Code, nextCalled
}

func TestIntentTokenRejectedOutsidePostGroceries(t *testing.T) {
	cases := []struct{ method, path string }{
		{http.MethodGet, "/grocery-lists"},
		{http.MethodGet, "/groceries"},
		{http.MethodDelete, "/groceries"},
		{http.MethodDelete, "/groceries/1"},
		{http.MethodPost, "/grocery-lists"},
		{http.MethodPost, "/auth/logout"},
		{http.MethodPost, "/auth/intent-token"},
		{http.MethodGet, "/guilds"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			repo := newFakeIntentTokenRepo()
			status, nextCalled := runAuth(t, repo, tc.method, tc.path, "Bearer "+validIntentToken)
			if status != http.StatusForbidden || nextCalled {
				t.Fatalf("got status %d nextCalled=%v, want 403 without next", status, nextCalled)
			}
			if repo.lookups != 0 {
				t.Fatalf("token should not be looked up for a disallowed route")
			}
		})
	}
}

func TestIntentTokenUnknownOrExpired(t *testing.T) {
	for _, token := range []string{"gbi_garbage", expiredIntentToken} {
		t.Run(token, func(t *testing.T) {
			repo := newFakeIntentTokenRepo()
			status, nextCalled := runAuth(t, repo, http.MethodPost, "/groceries", "Bearer "+token)
			if status != http.StatusUnauthorized || nextCalled {
				t.Fatalf("got status %d nextCalled=%v, want 401 without next", status, nextCalled)
			}
		})
	}
}

func TestIntentTokenValidRunsGuildCheck(t *testing.T) {
	repo := newFakeIntentTokenRepo()
	// No X-Guild-ID: the shared guild check runs and rejects with 400, the same as the JWT path.
	status, nextCalled := runAuth(t, repo, http.MethodPost, "/groceries", "Bearer "+validIntentToken)
	if status != http.StatusBadRequest || nextCalled {
		t.Fatalf("got status %d nextCalled=%v, want 400 without next", status, nextCalled)
	}
	if repo.lookups != 1 {
		t.Fatalf("expected one token lookup, got %d", repo.lookups)
	}
}

func TestAuthPathExemptions(t *testing.T) {
	repo := newFakeIntentTokenRepo()
	for _, path := range []string{"/auth/intent-token", "/auth/logout"} {
		status, nextCalled := runAuth(t, repo, http.MethodPost, path, "")
		if status != http.StatusUnauthorized || nextCalled {
			t.Fatalf("%s: got status %d nextCalled=%v, want 401 without next", path, status, nextCalled)
		}
	}
	for _, path := range []string{"/auth/token", "/auth/refresh"} {
		status, nextCalled := runAuth(t, repo, http.MethodPost, path, "")
		if status != http.StatusOK || !nextCalled {
			t.Fatalf("%s: got status %d nextCalled=%v, want pass-through", path, status, nextCalled)
		}
	}
}
