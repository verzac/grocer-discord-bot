package native

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
)

const (
	ContentUseEphemeralDescription      = "Enable ephemeral message replies from GroceryBot, which are only visible to you and will disappear."
	ContentUseGrobulkReplaceDescription = "If enabled, using /grobulk replaces the existing items in your list instead of adding new ones."
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

	// get config
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

	subCommand := options[0]
	optionNameToOptionsMapping := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(subCommand.Options))
	for _, option := range subCommand.Options {
		optionNameToOptionsMapping[option.Name] = option
	}

	switch subCommand.Name {
	case "set":
		setConfig(c, config, optionNameToOptionsMapping)
	case "get":
		getConfig(c, config)
	default:
		c.onError(errors.New("unknown subcommand"))
		return
	}

}

func getConfig(c *NativeSlashHandlingContext, config *models.GuildConfig) {
	enabledStr := func(enabled bool) string {
		if enabled {
			return "✅ ON"
		}
		return "❌ OFF"
	}
	message := fmt.Sprintf(`
# 🔨 Configuration
- **Use ephemeral**: %s - %s
- **Use grobulk replace**: %s - %s
`,
		enabledStr(config.UseEphemeral), ContentUseEphemeralDescription,
		enabledStr(!config.UseGrobulkAppend), ContentUseGrobulkReplaceDescription)

	c.reply(strings.TrimSpace(message))
}

func setConfig(c *NativeSlashHandlingContext, existingConfig *models.GuildConfig, optionNameToOptionsMapping map[string]*discordgo.ApplicationCommandInteractionDataOption) {
	newConfig := *existingConfig // copy

	// set all options
	if useEphemeral, ok := optionNameToOptionsMapping["use_ephemeral"]; ok && useEphemeral != nil {
		newConfig.UseEphemeral = useEphemeral.BoolValue()
	}

	if useGrobulkReplace, ok := optionNameToOptionsMapping["use_grobulk_replace"]; ok && useGrobulkReplace != nil {
		newConfig.UseGrobulkAppend = !useGrobulkReplace.BoolValue()
	}

	// save
	if err := c.guildConfigRepository.Put(&newConfig); err != nil {
		c.onError(err)
		return
	}

	c.reply("Configuration updated.")
}
