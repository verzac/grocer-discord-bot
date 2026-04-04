package api

import (
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type guildListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type guildsHandlerDeps struct {
	logger          *zap.Logger
	userSessionRepo repositories.UserSessionRepository
	oauth2Cfg       *oauth2.Config
	encryptionKey   string
	botSession      *discordgo.Session
}

func (d *guildsHandlerDeps) handleGuilds(c echo.Context) error {
	ctx := c.Request().Context()
	ac, ok := c.(*AuthContext)
	if !ok || ac.UserID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Bearer authentication required.")
	}
	row, err := d.userSessionRepo.GetByDiscordUserID(ctx, ac.UserID)
	if err != nil {
		return err
	}
	if row == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Session not found; please sign in again.")
	}
	access, err := auth.DecryptWithKey(d.encryptionKey, row.DiscordAccessToken)
	if err != nil {
		return err
	}
	refresh, err := auth.DecryptWithKey(d.encryptionKey, row.DiscordRefreshToken)
	if err != nil {
		return err
	}
	oauthTok := &oauth2.Token{
		AccessToken:  access,
		RefreshToken: refresh,
		Expiry:       row.DiscordTokenExpiry,
	}
	if time.Until(row.DiscordTokenExpiry) < 2*time.Minute {
		newTok, rerr := refreshDiscordOAuthToken(ctx, d.oauth2Cfg, refresh)
		if rerr != nil {
			d.logger.Warn("Discord OAuth refresh failed", zap.Error(rerr))
			return echo.NewHTTPError(http.StatusUnauthorized, "Discord session expired; please sign in again.")
		}
		encAccess, encErr := auth.EncryptWithKey(d.encryptionKey, newTok.AccessToken)
		if encErr != nil {
			return encErr
		}
		encRef := row.DiscordRefreshToken
		if newTok.RefreshToken != "" {
			encRef, encErr = auth.EncryptWithKey(d.encryptionKey, newTok.RefreshToken)
			if encErr != nil {
				return encErr
			}
		}
		row.DiscordAccessToken = encAccess
		row.DiscordRefreshToken = encRef
		row.DiscordTokenExpiry = discordTokenExpiry(newTok)
		if err := d.userSessionRepo.UpsertByDiscordUserID(ctx, row); err != nil {
			return err
		}
		oauthTok = newTok
	}
	userGuilds, err := fetchDiscordUserGuilds(ctx, oauthTok.AccessToken)
	if err != nil {
		newTok, rerr := refreshDiscordOAuthToken(ctx, d.oauth2Cfg, refresh)
		if rerr != nil {
			d.logger.Warn("Discord guild fetch and refresh failed", zap.Error(err), zap.NamedError("refresh", rerr))
			return echo.NewHTTPError(http.StatusUnauthorized, "Discord session expired; please sign in again.")
		}
		encAccess, encErr := auth.EncryptWithKey(d.encryptionKey, newTok.AccessToken)
		if encErr != nil {
			return encErr
		}
		encRef := row.DiscordRefreshToken
		if newTok.RefreshToken != "" {
			encRef, encErr = auth.EncryptWithKey(d.encryptionKey, newTok.RefreshToken)
			if encErr != nil {
				return encErr
			}
		}
		row.DiscordAccessToken = encAccess
		row.DiscordRefreshToken = encRef
		row.DiscordTokenExpiry = discordTokenExpiry(newTok)
		if err := d.userSessionRepo.UpsertByDiscordUserID(ctx, row); err != nil {
			return err
		}
		userGuilds, err = fetchDiscordUserGuilds(ctx, newTok.AccessToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, "Failed to load Discord guilds.")
		}
	}
	botGuildIDs := make(map[string]struct{})
	if d.botSession != nil {
		for _, g := range d.botSession.State.Guilds {
			if g != nil {
				botGuildIDs[g.ID] = struct{}{}
			}
		}
	}
	out := make([]guildListItem, 0)
	for _, ug := range userGuilds {
		if _, ok := botGuildIDs[ug.ID]; !ok {
			continue
		}
		icon := ""
		if ug.Icon != "" {
			icon = ug.Icon
		}
		out = append(out, guildListItem{ID: ug.ID, Name: ug.Name, Icon: icon})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"guilds": out})
}
