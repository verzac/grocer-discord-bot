package routeauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/auth"
	apimw "github.com/verzac/grocer-discord-bot/handlers/api/middleware"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"github.com/verzac/grocer-discord-bot/services/demo"
	"go.uber.org/zap"
)

type fakeIntentTokenRepo struct {
	deletedFor []string
	created    []*models.IntentToken
}

var _ repositories.IntentTokenRepository = &fakeIntentTokenRepo{}

func (f *fakeIntentTokenRepo) WithContext(ctx context.Context) repositories.IntentTokenRepository {
	return f
}

func (f *fakeIntentTokenRepo) Create(ctx context.Context, token *models.IntentToken) error {
	f.created = append(f.created, token)
	return nil
}

func (f *fakeIntentTokenRepo) FindByTokenHash(ctx context.Context, hash string) (*models.IntentToken, error) {
	return nil, nil
}

func (f *fakeIntentTokenRepo) DeleteByDiscordUserID(ctx context.Context, discordUserID string) error {
	f.deletedFor = append(f.deletedFor, discordUserID)
	return nil
}

const demoUserID = "demo-user"

// mintIntentToken calls POST /auth/intent-token with an AuthContext built from userID and isIntentToken,
// standing in for AuthMiddleware.
func mintIntentToken(t *testing.T, repo *fakeIntentTokenRepo, userID string, isIntentToken bool) *httptest.ResponseRecorder {
	t.Helper()
	t.Setenv("DEMO_USER_DISCORD_ID", demoUserID)
	t.Setenv("DEMO_LOGIN_PASSWORD", "pw")
	demo.Init(zap.NewNop(), nil)
	t.Cleanup(func() { demo.Service = nil })

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(&apimw.AuthContext{Context: c, UserID: userID, IsIntentToken: isIntentToken})
		}
	})
	Register(e, zap.NewNop(), &auth.OAuthSetup{}, nil, repo)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/auth/intent-token", nil))
	return rec
}

func TestMintIntentToken(t *testing.T) {
	repo := &fakeIntentTokenRepo{}
	rec := mintIntentToken(t, repo, "user-1", false)
	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	var body intentTokenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(body.IntentToken, auth.IntentTokenPrefix) || body.ExpiresIn != 2592000 {
		t.Fatalf("unexpected body: %+v", body)
	}
	if len(repo.deletedFor) != 1 || repo.deletedFor[0] != "user-1" {
		t.Fatalf("expected prior tokens deleted for user-1, got %v", repo.deletedFor)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected one token created, got %d", len(repo.created))
	}
	it := repo.created[0]
	if it.DiscordUserID != "user-1" || it.TokenHash != auth.HashIntentToken(body.IntentToken) {
		t.Fatalf("stored token does not match response: %+v", it)
	}
	if d := time.Until(it.ExpiresAt) - auth.DefaultIntentTokenTTL; d > time.Minute || d < -time.Minute {
		t.Fatalf("unexpected expiry %v", it.ExpiresAt)
	}
}

func TestMintIntentTokenForbidden(t *testing.T) {
	cases := map[string]struct {
		userID        string
		isIntentToken bool
	}{
		"demo user":    {demoUserID, false},
		"intent token": {"user-1", true},
		"basic auth":   {"", false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeIntentTokenRepo{}
			rec := mintIntentToken(t, repo, tc.userID, tc.isIntentToken)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
			}
			if len(repo.deletedFor) != 0 || len(repo.created) != 0 {
				t.Fatalf("repository must not be touched: %+v", repo)
			}
		})
	}
}
