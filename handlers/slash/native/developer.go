package native

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

const customIDGenerateNewApiKey = "generate_new_api_key"

var handleDeveloper NativeSlashHandler = func(c *NativeSlashHandlingContext) {
	if c.i.Member == nil {
		c.reply("This command can only be used in a server (to generate your API keys).")
		return
	}
	// check for perms
	userPermissions := c.i.Member.Permissions
	if userPermissions&discordgo.PermissionAdministrator != discordgo.PermissionAdministrator {
		c.reply("This command can only be used by someone with the Administrator permission in your server.")
		return
	}
	// check for existing API key
	guildID := c.i.GuildID
	keys, err := c.apiKeyRepository.FindApiKeysByGuildID(guildID)
	if err != nil {
		c.onError(err)
		return
	}
	// if exist then reconfirm whether or not they'd like to purge the old one
	if len(keys) >= 1 {
		if len(keys) > 1 {
			c.logger.Error("Non-fatal: Multiple API keys detected.", zap.String("ScopeGuildID", guildID))
		}
		// reconfirm - are they sure they want to replace the API key?
		// API KEYS CANNOT BE SHOWN MORE THAN ONCE
		if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You already have an API key for this server. Would you like to replace your existing one with a new one?",
				// Flags:   discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    customIDGenerateNewApiKey,
								Placeholder: "Create new API key?",
								Options: []discordgo.SelectMenuOption{
									{
										Label:       "Yes",
										Description: "Keep in mind that your old API key will stop working!",
										Value:       "yes",
									},
									{
										Label: "No - take me back!",
										Value: "no",
									},
								},
							},
						},
					},
				},
			},
		}); err != nil {
			c.onError(err)
			return
		}
	} else {
		// create new one - proxy to new api key creation handler
		createNewAPIKey(c)
		return
	}
}

var handleDeveloperCreateNewApiKey NativeSlashHandler = func(c *NativeSlashHandlingContext) {
	data := c.i.MessageComponentData()
	values := data.Values
	if len(values) != 1 {
		c.onError(fmt.Errorf("create_new_api_key handler: len(values) != 1, values: %+v", values))
		return
	}
	if values[0] == "yes" {
		createNewAPIKey(c)
		return
	} else {
		c.reply("Got it - no problem! No new API key has been created. Your old API key should still work!")
		return
	}
}

func createNewAPIKey(c *NativeSlashHandlingContext) {
	// create new one
	newApiKeyStr, err := auth.GenerateApiKey()
	if err != nil {
		c.onError(err)
		return
	}
	hashedNewApiKeyStr, err := auth.HashApiKey(newApiKeyStr)
	if err != nil {
		c.onError(err)
		return
	}
	newApiKey := &models.ApiKey{
		ApiKeyHashed: hashedNewApiKeyStr,
		CreatedByID:  c.i.Member.User.ID,
		Scope:        c.apiKeyRepository.GetScopeForGuild(c.i.GuildID),
	}
	if err := c.apiKeyRepository.Put(newApiKey); err != nil {
		c.onError(err)
		return
	}
	c.replyWithOption(fmt.Sprintf("Yay, we've generated a new API key for you! Here is your API key:\n`%s`", newApiKeyStr), replyOptions{IsPrivate: true})
}
