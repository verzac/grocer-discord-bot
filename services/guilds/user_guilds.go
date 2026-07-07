package guilds

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	cache "github.com/patrickmn/go-cache"
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/services/oauthsession"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

var discordUsersMeGuildsURL_ = "https://discord.com/api/users/@me/guilds"

var (
	userGuildsCache = cache.New(60*time.Second, 2*time.Minute)
	userGuildsSF    singleflight.Group
)

func (s *GuildsServiceImpl) GetUserGuilds(ctx context.Context, userID string) ([]dto.UserGuild, error) {
	if cached, ok := userGuildsCache.Get(userID); ok {
		return cached.([]dto.UserGuild), nil
	}

	v, err, _ := userGuildsSF.Do(userID, func() (any, error) {
		guilds, err := s.fetchUserGuilds(ctx, userID)
		if err != nil {
			return nil, err
		}
		userGuildsCache.Set(userID, guilds, cache.DefaultExpiration)
		return guilds, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]dto.UserGuild), nil
}

func (s *GuildsServiceImpl) fetchUserGuilds(ctx context.Context, userID string) ([]dto.UserGuild, error) {
	client, err := oauthsession.Service.DiscordUserHTTPClient(ctx, userID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discordUsersMeGuildsURL_, nil)
	if err != nil {
		s.logger.Error("build @me/guilds request", zap.Error(err))
		return nil, echo.NewHTTPError(502, "Could not reach Discord.")
	}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("discord @me/guilds request failed", zap.Error(err))
		return nil, echo.NewHTTPError(502, "Could not reach Discord.")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, echo.NewHTTPError(401, "Discord session expired; please re-authenticate.")
	}
	if resp.StatusCode == 429 {
		s.logger.Warn("discord @me/guilds rate limited",
			zap.String("userID", userID),
			zap.String("retryAfter", resp.Header.Get("Retry-After")),
		)
		return nil, echo.NewHTTPError(429, "Discord rate limit exceeded; please try again shortly.")
	}
	if resp.StatusCode != 200 {
		s.logger.Warn("discord @me/guilds non-OK", zap.Int("status", resp.StatusCode))
		return nil, echo.NewHTTPError(502, "Could not load guilds from Discord.")
	}

	var raw []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Icon string `json:"icon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		s.logger.Error("decode discord guilds", zap.Error(err))
		return nil, echo.NewHTTPError(502, "Could not load guilds from Discord.")
	}

	guilds := make([]dto.UserGuild, len(raw))
	for i, g := range raw {
		guilds[i] = dto.UserGuild{ID: g.ID, Name: g.Name, Icon: g.Icon}
	}
	return guilds, nil
}
