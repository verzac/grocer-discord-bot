package native

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
)

var handleConfig NativeSlashHandler = func(c *NativeSlashHandlingContext) {
	if c.i.Member == nil {
		c.reply("This command can only be used in a server (since configurations are stored for each server).")
		return
	}
	userPermissions := c.i.Member.Permissions
	if userPermissions&discordgo.PermissionAdministrator != discordgo.PermissionAdministrator {
		c.reply("This command can only be used by people with the Administrator permission in your server.")
		return
	}
	guildID := c.i.GuildID
	options := c.i.ApplicationCommandData().Options
	if len(options) != 1 || options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		c.onError(errMissingSubcommand)
		return
	}
	subCommand := options[0]
	if subCommand.Name != "set" {
		c.onError(errors.New("unknown subcommand"))
		return
	}
	optionNameToOptionsMapping := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(subCommand.Options))
	for _, option := range subCommand.Options {
		optionNameToOptionsMapping[option.Name] = option
	}

	// get config to set
	config, err := c.guildConfigRepository.Get(guildID)
	if err != nil {
		c.onError(err)
		return
	}
	if config == nil {
		config = &models.GuildConfig{
			GuildID: guildID,
		}
	}

	// set all options
	if useEphemeral, ok := optionNameToOptionsMapping["use_ephemeral"]; ok && useEphemeral != nil {
		config.UseEphemeral = useEphemeral.BoolValue()
	}

	// save
	if err := c.guildConfigRepository.Put(config); err != nil {
		c.onError(err)
		return
	}

	c.reply("Configuration saved.")
}
