package repositories_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/verzac/grocer-discord-bot/db"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newMigratedDB returns a fresh SQLite database with all of db/changelog applied.
func newMigratedDB(t *testing.T) *gorm.DB {
	t.Helper()
	changelog, err := filepath.Abs(filepath.Join("..", "db", "changelog"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GROCER_BOT_DB_SOURCE_CHANGELOG", "file://"+changelog)
	g, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(g, zap.NewNop(), "local"); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestIntentTokenRepository(t *testing.T) {
	g := newMigratedDB(t)
	r := &repositories.IntentTokenRepositoryImpl{DB: g}
	ctx := context.Background()
	exp := time.Now().UTC().Add(time.Hour).Truncate(time.Second)

	upsert := func(userID, hash string, expiresAt time.Time) error {
		return r.Upsert(ctx, &models.IntentToken{DiscordUserID: userID, TokenHash: hash, ExpiresAt: expiresAt})
	}
	countFor := func(userID string) int64 {
		var n int64
		if err := g.Model(&models.IntentToken{}).Where("discord_user_id = ?", userID).Count(&n).Error; err != nil {
			t.Fatal(err)
		}
		return n
	}

	if err := upsert("u1", "h1", exp); err != nil {
		t.Fatal(err)
	}
	if err := upsert("u1", "h2", exp.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if n := countFor("u1"); n != 1 {
		t.Fatalf("expected one token for u1 after re-mint, got %d", n)
	}
	if it, err := r.FindByTokenHash(ctx, "h1"); err != nil || it != nil {
		t.Fatalf("replaced token must not be found: %+v %v", it, err)
	}
	it, err := r.FindByTokenHash(ctx, "h2")
	if err != nil || it == nil {
		t.Fatalf("latest token not found: %v", err)
	}
	if it.DiscordUserID != "u1" || !it.ExpiresAt.Equal(exp.Add(time.Hour)) {
		t.Fatalf("latest token not stored correctly: %+v", it)
	}

	if err := upsert("u2", "h2", exp); err == nil {
		t.Fatal("expected a duplicate token_hash for another user to be rejected")
	}

	if err := upsert("u2", "h3", exp); err != nil {
		t.Fatal(err)
	}
	if err := r.DeleteByDiscordUserID(ctx, "u1"); err != nil {
		t.Fatal(err)
	}
	if n := countFor("u1"); n != 0 {
		t.Fatalf("expected u1 tokens deleted, got %d", n)
	}
	if n := countFor("u2"); n != 1 {
		t.Fatalf("deleting u1 must not affect u2, got %d", n)
	}
	if it, err := r.FindByTokenHash(ctx, "missing"); err != nil || it != nil {
		t.Fatalf("missing token should return nil, nil: %+v %v", it, err)
	}
}
