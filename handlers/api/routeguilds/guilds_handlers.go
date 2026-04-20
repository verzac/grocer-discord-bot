package routeguilds

import (
	"encoding/json"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/dto"
	apimw "github.com/verzac/grocer-discord-bot/handlers/api/middleware"
	"github.com/verzac/grocer-discord-bot/services/oauthsession"
	"go.uber.org/zap"
)

const discordUsersMeGuildsURL = "https://discord.com/api/users/@me/guilds"

type discordUserGuild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// Register mounts GET /guilds (Bearer JWT; no X-Guild-ID). Requires oauthsession.Init before startup.
func Register(e *echo.Echo, logger *zap.Logger, discordSess *discordgo.Session) {
	logger = logger.Named("guilds")

	e.GET("/guilds", func(c echo.Context) error {
		ctx := c.Request().Context()
		authContext := c.(*apimw.AuthContext)
		userID := authContext.UserID
		if userID == "" {
			return echo.NewHTTPError(401, "Session not found; please re-authenticate.")
		}

		client, err := oauthsession.Service.DiscordUserHTTPClient(ctx, userID)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, discordUsersMeGuildsURL, nil)
		if err != nil {
			logger.Error("build @me/guilds request", zap.Error(err))
			return echo.NewHTTPError(502, "Could not reach Discord.")
		}
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("discord @me/guilds request failed", zap.Error(err))
			return echo.NewHTTPError(502, "Could not reach Discord.")
		}
		defer resp.Body.Close()
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return echo.NewHTTPError(401, "Discord session expired; please re-authenticate.")
		}
		if resp.StatusCode != 200 {
			logger.Warn("discord @me/guilds non-OK", zap.Int("status", resp.StatusCode))
			return echo.NewHTTPError(502, "Could not load guilds from Discord.")
		}

		var userGuilds []discordUserGuild
		if err := json.NewDecoder(resp.Body).Decode(&userGuilds); err != nil {
			logger.Error("decode discord guilds", zap.Error(err))
			return echo.NewHTTPError(502, "Could not load guilds from Discord.")
		}

		if discordSess == nil {
			return echo.NewHTTPError(500, "Cannot resolve bot guilds.")
		}
		botGuildIDs := make(map[string]struct{}, len(discordSess.State.Guilds))
		for _, g := range discordSess.State.Guilds {
			if g != nil {
				botGuildIDs[g.ID] = struct{}{}
			}
		}

		out := dto.UserGuildsResponse{Guilds: make([]dto.UserGuild, 0, len(userGuilds))}
		for _, ug := range userGuilds {
			if _, ok := botGuildIDs[ug.ID]; ok {
				out.Guilds = append(out.Guilds, dto.UserGuild{
					ID:   ug.ID,
					Name: ug.Name,
					Icon: ug.Icon,
				})
			}
		}
		return c.JSON(200, out)
	})
}
