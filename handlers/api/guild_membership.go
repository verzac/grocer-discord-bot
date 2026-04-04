package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	cache "github.com/patrickmn/go-cache"
)

var guildMemberVerifyCache = cache.New(60*time.Second, 2*time.Minute)

func pathRequiresGuildIDHeader(method, path string) bool {
	switch {
	case path == "/grocery-lists" && method == http.MethodGet:
		return true
	case path == "/registrations" && method == http.MethodGet:
		return true
	case path == "/groceries" && method == http.MethodPost:
		return true
	case method == http.MethodDelete && strings.HasPrefix(path, "/groceries/") && len(path) > len("/groceries/"):
		return true
	default:
		return false
	}
}

func verifyUserGuildMembership(sess *discordgo.Session, guildID, userID string) error {
	if sess == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Discord session not available.")
	}
	cacheKey := guildID + ":" + userID
	if _, ok := guildMemberVerifyCache.Get(cacheKey); ok {
		return nil
	}
	if _, err := sess.State.Guild(guildID); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Bot is not in this guild.")
	}
	if _, err := sess.GuildMember(guildID, userID); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Not a member of this guild.")
	}
	guildMemberVerifyCache.Set(cacheKey, true, cache.DefaultExpiration)
	return nil
}
