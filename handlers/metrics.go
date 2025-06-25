package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/monitoring/groprometheus"
)

func GetServerCountFromSession(s *discordgo.Session) int {
	return len(s.State.Guilds)
}

func UpdateServerCountMetric(s *discordgo.Session) int {
	serverCount := GetServerCountFromSession(s)
	groprometheus.UpdateDiscordServers(serverCount)
	return serverCount
}
