package routeguilds

import (
	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/dto"
	apimw "github.com/verzac/grocer-discord-bot/handlers/api/middleware"
	"github.com/verzac/grocer-discord-bot/services/guilds"
)

// Register mounts GET /guilds (Bearer JWT; no X-Guild-ID). Requires oauthsession.Init before startup.
func Register(e *echo.Echo, discordSess *discordgo.Session) {
	e.GET("/guilds", func(c echo.Context) error {
		authContext := c.(*apimw.AuthContext)
		userID := authContext.UserID
		if userID == "" {
			return echo.NewHTTPError(401, "Session not found; please re-authenticate.")
		}

		userGuilds, err := guilds.Service.GetUserGuilds(c.Request().Context(), userID)
		if err != nil {
			return err
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
				out.Guilds = append(out.Guilds, ug)
			}
		}
		return c.JSON(200, out)
	})
}
