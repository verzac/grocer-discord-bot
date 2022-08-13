package native

import (
	"encoding/base64"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/auth"
	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

const customIDGenerateNewApiClient = "generate_new_api_client"

var handleDeveloper NativeSlashHandler = func(c *NativeSlashHandlingContext) {
	if c.i.Member == nil {
		c.reply("This command can only be used in a server (to generate your API Client).")
		return
	}
	// check for perms
	userPermissions := c.i.Member.Permissions
	if userPermissions&discordgo.PermissionAdministrator != discordgo.PermissionAdministrator {
		c.reply("This command can only be used by people with the Administrator permission in your server.")
		return
	}
	// check for existing API client
	guildID := c.i.GuildID
	clients, err := c.apiClientRepository.FindApiClientsByGuildID(guildID)
	if err != nil {
		c.onError(err)
		return
	}
	// if exist then reconfirm whether or not they'd like to purge the old one
	if len(clients) >= 1 {
		if len(clients) > 1 {
			c.logger.Error("Non-fatal: Multiple API clients detected.", zap.String("ScopeGuildID", guildID))
		}
		// reconfirm - are they sure they want to replace the API client?
		// API CREDENTIALS CANNOT BE SHOWN MORE THAN ONCE
		if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You already have an API Client ID & Secret for this server. Would you like to replace your existing one with a new one?",
				Flags:   discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    customIDGenerateNewApiClient,
								Placeholder: "Create new API Client?",
								Options: []discordgo.SelectMenuOption{
									{
										Label:       "Yes",
										Description: "Keep in mind that your old Client ID & Client Secret will stop working!",
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
		// create new one - proxy to new api client creation handler
		createNewApiClient(c)
		return
	}
}

var handleDeveloperCreateNewApiClient NativeSlashHandler = func(c *NativeSlashHandlingContext) {
	data := c.i.MessageComponentData()
	values := data.Values
	if len(values) != 1 {
		c.onError(fmt.Errorf("create_new_api_client handler: len(values) != 1, values: %+v", values))
		return
	}
	if values[0] == "yes" {
		createNewApiClient(c)
		return
	} else {
		c.reply("Got it - no problem! No new API Client has been created. Your old API Client ID & Client Secret should still work!")
		return
	}
}

func createNewApiClient(c *NativeSlashHandlingContext) {
	// create new one
	clientSecret, err := auth.GenerateKey()
	if err != nil {
		c.onError(err)
		return
	}
	hashedNewClientSecret, err := auth.HashKey(clientSecret)
	if err != nil {
		c.onError(err)
		return
	}
	clientID := c.i.GuildID
	newApiClient := &models.ApiClient{
		ClientID:     clientID,
		ClientSecret: hashedNewClientSecret,
		CreatedByID:  c.i.Member.User.ID,
		Scope:        auth.GetScopeForGuild(c.i.GuildID),
	}
	if err := c.apiClientRepository.DeleteAllByClientID(clientID); err != nil {
		c.onError(err)
		return
	}
	if err := c.apiClientRepository.Put(newApiClient); err != nil {
		c.onError(err)
		return
	}
	err = c.replyWithOption(fmt.Sprintf(`
Yay, we've generated a new API Client for you! Here's the deets:
`+"```"+`
Client ID: %s
Client Secret: %s
`+"```"+`

You can use your API Client by combining the two keys above into a Basic auth header using the following format `+"`"+`client_id:client_secret`+"`"+` with Base-64 encoding.

e.g.
`+"```"+`
Authorization: Basic %s
`+"```"+`
*Please store this somewhere safe!* We can't retrieve this at a later time - if you lose these you'd have to re-generate your API client by running `+"`"+`/developer`+"`"+` again.
`, clientID, clientSecret, base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))), replyOptions{IsPrivate: true})
	if err != nil {
		c.onError(err)
		return
	}

	followUpMsg := fmt.Sprintf(":wave: Heyo! Just letting yall know that %s has just created an API Client for this server. This means that their programs / applications would be able to access GroceryBot's data for this server. Thank you, and have a nice day!", c.i.Member.Mention())
	if _, err := c.s.FollowupMessageCreate(c.i.Interaction, true, &discordgo.WebhookParams{
		Content: followUpMsg,
	}); err != nil {
		c.onError(err)
		return
	}
}
