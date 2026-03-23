package modal

import (
	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
)

func getGroremoveCommandContext(i *discordgo.InteractionCreate, commandName string) (*handlers.CommandContext, error) {
	data := i.ModalSubmitData()
	argStr := data.Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).
		Value
	return &handlers.CommandContext{
		Command:                     "!" + commandName,
		GrocerySublist:              "",
		ArgStr:                      argStr,
		GuildID:                     i.GuildID,
		AuthorID:                    i.Member.User.ID,
		ChannelID:                   i.ChannelID,
		AuthorUsername:              i.Member.User.Username,
		AuthorUsernameDiscriminator: i.Member.User.Discriminator,
		CommandSourceType:           handlers.CommandSourceSlashCommand,
		Interaction:                 i.Interaction,
	}, nil
}

func handleGroremoveCommand(c *ModalCreationContext) (*discordgo.InteractionResponseData, error) {
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "groremove_input",
					Label:       "Items to remove",
					Placeholder: "e.g. 1 2 or chicken katsu",
					Required:    true,
					MaxLength:   4000,
					Style:       discordgo.TextInputParagraph,
				},
			},
		},
	}
	return &discordgo.InteractionResponseData{
		CustomID:   "groremove",
		Title:      "!groremove",
		Components: components,
	}, nil
}
