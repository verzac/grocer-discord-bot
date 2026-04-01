package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const discordAPIBase = "https://discord.com/api/v10"

type discordUserMe struct {
	ID string `json:"id"`
}

type discordGuildPartial struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func fetchDiscordUserMe(ctx context.Context, oauthAccessToken string) (*discordUserMe, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discordAPIBase+"/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+oauthAccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("discord users/@me: status %d: %s", resp.StatusCode, string(body))
	}
	var out discordUserMe
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func fetchDiscordUserGuilds(ctx context.Context, oauthAccessToken string) ([]discordGuildPartial, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discordAPIBase+"/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+oauthAccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("discord users/@me/guilds: status %d: %s", resp.StatusCode, string(body))
	}
	var out []discordGuildPartial
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func refreshDiscordOAuthToken(ctx context.Context, cfg *oauth2.Config, refreshToken string) (*oauth2.Token, error) {
	if cfg == nil || refreshToken == "" {
		return nil, fmt.Errorf("missing oauth config or refresh token")
	}
	return cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
}

func discordTokenExpiry(t *oauth2.Token) time.Time {
	if t == nil {
		return time.Time{}
	}
	if !t.Expiry.IsZero() {
		return t.Expiry
	}
	return time.Now().Add(time.Hour)
}
