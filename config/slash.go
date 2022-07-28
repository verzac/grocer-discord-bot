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

var (
	disabledSlashCommands = []string{
		"grobulk",
	}
)

func GetIgnoredSlashCommands(grobotVersion string) map[string]interface{} {
	ignoredSlashCommands := make(map[string]interface{}, 0)
	if grobotVersion != "local" {
		for _, cmd := range disabledSlashCommands {
			ignoredSlashCommands[cmd] = struct{}{}
		}
	}
	ignoredSlashCommandsStr := os.Getenv("GROCER_BOT_SLASH_IGNORE_COMMANDS")
	if ignoredSlashCommandsStr != "" {
		ignoredSlashCommandsFromEnv := strings.Split(ignoredSlashCommandsStr, ",")
		for _, cmd := range ignoredSlashCommandsFromEnv {
			ignoredSlashCommands[cmd] = struct{}{}
		}
	}
	return ignoredSlashCommands
}
