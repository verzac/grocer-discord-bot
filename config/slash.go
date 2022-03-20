package config

import (
	"os"
	"strings"
)

func GetGuildIDsToRegisterSlashCommandsOn() []string {
	defaultGuildIDs := []string{
		"",
	}
	guildIDsOverride := os.Getenv("GROCER_BOT_SLASH_REGISTER_GUILD_ID_LIST")
	if len(guildIDsOverride) > 0 {
		return strings.Split(guildIDsOverride, ",")
	}
	return defaultGuildIDs
}

func GetGuildIDsToDeregisterSlashCommandsFrom() []string {
	defaultGuildIDs := []string{}
	guildIDsOverride := os.Getenv("GROCER_BOT_SLASH_DEREGISTER_GUILD_ID_LIST")
	if len(guildIDsOverride) > 0 {
		return strings.Split(guildIDsOverride, ",")
	}
	return defaultGuildIDs
}
