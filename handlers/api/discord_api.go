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

const discordAPIUsersMe = "https://discord.com/api/users/@me"
const discordAPIMyGuilds = "https://discord.com/api/users/@me/guilds"

type discordUserMe struct {
	ID string `json:"id"`
}

type discordGuildPartial struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Icon *string `json:"icon"`
}

func fetchDiscordUserMe(ctx context.Context, accessToken string, refreshToken string, expiry *time.Time) (*discordUserMe, *oauth2.Token, error) {
	cfg := discordOAuthConfig()
	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if expiry != nil {
		tok.Expiry = *expiry
	}
	ts := cfg.TokenSource(ctx, tok)
	client := oauth2.NewClient(ctx, ts)
	resp, err := client.Get(discordAPIUsersMe)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("discord /users/@me: status %d: %s", resp.StatusCode, string(body))
	}
	var u discordUserMe
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, nil, err
	}
	newTok, err := ts.Token()
	if err != nil {
		newTok = tok
	}
	return &u, newTok, nil
}

func fetchDiscordMyGuilds(ctx context.Context, accessToken string, refreshToken string, expiry *time.Time) ([]discordGuildPartial, *oauth2.Token, error) {
	cfg := discordOAuthConfig()
	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if expiry != nil {
		tok.Expiry = *expiry
	}
	ts := cfg.TokenSource(ctx, tok)
	client := oauth2.NewClient(ctx, ts)
	resp, err := client.Get(discordAPIMyGuilds)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("discord /users/@me/guilds: status %d: %s", resp.StatusCode, string(body))
	}
	var guilds []discordGuildPartial
	if err := json.Unmarshal(body, &guilds); err != nil {
		return nil, nil, err
	}
	newTok, err := ts.Token()
	if err != nil {
		newTok = tok
	}
	return guilds, newTok, nil
}
