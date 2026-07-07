package guilds

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/services/oauthsession"
	"go.uber.org/zap"
)

type mockOAuthService struct {
	client *http.Client
}

func (m *mockOAuthService) DiscordUserHTTPClient(_ context.Context, _ string) (*http.Client, error) {
	return m.client, nil
}

func setupTest(t *testing.T, handler http.HandlerFunc) *GuildsServiceImpl {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	origURL := discordUsersMeGuildsURL_
	discordUsersMeGuildsURL_ = srv.URL
	t.Cleanup(func() { discordUsersMeGuildsURL_ = origURL })

	oauthsession.Service = &mockOAuthService{client: srv.Client()}
	t.Cleanup(func() { oauthsession.Service = nil })

	userGuildsCache.Flush()

	logger, _ := zap.NewDevelopment()
	return &GuildsServiceImpl{logger: logger}
}

func TestGetUserGuilds_Success(t *testing.T) {
	var callCount atomic.Int32
	svc := setupTest(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{
			{"id": "1", "name": "Guild One", "icon": "icon1"},
			{"id": "2", "name": "Guild Two", "icon": "icon2"},
		})
	})

	guilds, err := svc.GetUserGuilds(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(guilds) != 2 {
		t.Fatalf("expected 2 guilds, got %d", len(guilds))
	}
	if guilds[0].ID != "1" || guilds[1].Name != "Guild Two" {
		t.Fatalf("unexpected guild data: %+v", guilds)
	}
	if callCount.Load() != 1 {
		t.Fatalf("expected 1 Discord call, got %d", callCount.Load())
	}
}

func TestGetUserGuilds_CacheHit(t *testing.T) {
	var callCount atomic.Int32
	svc := setupTest(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{
			{"id": "1", "name": "Guild One", "icon": "icon1"},
		})
	})

	// First call populates cache.
	_, err := svc.GetUserGuilds(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call should hit cache.
	guilds, err := svc.GetUserGuilds(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(guilds) != 1 {
		t.Fatalf("expected 1 guild from cache, got %d", len(guilds))
	}
	if callCount.Load() != 1 {
		t.Fatalf("expected 1 Discord call (cached), got %d", callCount.Load())
	}
}

func TestGetUserGuilds_DifferentUsersSeparateCaches(t *testing.T) {
	var callCount atomic.Int32
	svc := setupTest(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{
			{"id": "1", "name": "Guild", "icon": ""},
		})
	})

	_, _ = svc.GetUserGuilds(context.Background(), "user-1")
	_, _ = svc.GetUserGuilds(context.Background(), "user-2")

	if callCount.Load() != 2 {
		t.Fatalf("expected 2 Discord calls (different users), got %d", callCount.Load())
	}
}

func TestGetUserGuilds_Singleflight(t *testing.T) {
	var callCount atomic.Int32
	started := make(chan struct{})
	proceed := make(chan struct{})

	svc := setupTest(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		started <- struct{}{}
		<-proceed
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{
			{"id": "1", "name": "Guild", "icon": ""},
		})
	})

	var wg sync.WaitGroup
	results := make([][]dto.UserGuild, 3)
	errs := make([]error, 3)

	for i := range 3 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = svc.GetUserGuilds(context.Background(), "user-1")
		}(i)
	}

	// Wait for the first request to hit the server, then unblock.
	<-started
	close(proceed)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d got error: %v", i, err)
		}
	}
	if callCount.Load() != 1 {
		t.Fatalf("expected 1 Discord call (singleflight), got %d", callCount.Load())
	}
	for i, g := range results {
		if len(g) != 1 || g[0].ID != "1" {
			t.Fatalf("goroutine %d got unexpected result: %+v", i, g)
		}
	}
}

func TestGetUserGuilds_429NotCached(t *testing.T) {
	var callCount atomic.Int32
	svc := setupTest(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		if callCount.Load() == 1 {
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(429)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{
			{"id": "1", "name": "Guild", "icon": ""},
		})
	})

	_, err := svc.GetUserGuilds(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error on 429")
	}

	// Second call should retry (not return cached error).
	guilds, err := svc.GetUserGuilds(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected success on retry, got: %v", err)
	}
	if len(guilds) != 1 {
		t.Fatalf("expected 1 guild, got %d", len(guilds))
	}
	if callCount.Load() != 2 {
		t.Fatalf("expected 2 Discord calls, got %d", callCount.Load())
	}
}

func TestGetUserGuilds_401NotCached(t *testing.T) {
	var callCount atomic.Int32
	svc := setupTest(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		if callCount.Load() == 1 {
			w.WriteHeader(401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{
			{"id": "1", "name": "Guild", "icon": ""},
		})
	})

	_, err := svc.GetUserGuilds(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error on 401")
	}

	guilds, err := svc.GetUserGuilds(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected success on retry, got: %v", err)
	}
	if len(guilds) != 1 {
		t.Fatalf("expected 1 guild, got %d", len(guilds))
	}
	if callCount.Load() != 2 {
		t.Fatalf("expected 2 Discord calls, got %d", callCount.Load())
	}
}
