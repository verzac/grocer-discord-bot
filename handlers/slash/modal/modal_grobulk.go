package modal

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
)

func getGrobulkInputFromEntries(groceryEntries []models.GroceryEntry) string {
	msg := ""
	for _, grocery := range groceryEntries {
		msg += fmt.Sprintf("%s\n", grocery.ItemDesc)
	}
	return msg
}

func getGrobulkCommandContext(i *discordgo.InteractionCreate, commandName string) (*handlers.CommandContext, error) {
	data := i.ModalSubmitData()
	argStr := data.Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).
		Value
	return &handlers.CommandContext{
		Command:                     "!" + commandName,
		GrocerySublist:              "",
		ArgStr:                      argStr,
		GuildID:                     i.GuildID,
		AuthorID:                    i.Member.User.ID, // nil-check this in caller
		ChannelID:                   i.ChannelID,
		AuthorUsername:              i.Member.User.Username,
		AuthorUsernameDiscriminator: i.Member.User.Discriminator,
		CommandSourceType:           handlers.CommandSourceSlashCommand,
		Interaction:                 i.Interaction,
	}, nil
}

func handleGrobulkCommand(c *ModalCreationContext) (*discordgo.InteractionResponseData, error) {
	guildID := c.guildID
	guildConfig, err := c.guildConfigRepository.Get(guildID)
	if err != nil {
		return nil, err
	}
	if guildConfig == nil {
		guildConfig = &models.GuildConfig{
			GuildID: guildID,
		}
	}
	useGrobulkAppend := guildConfig.UseGrobulkAppend
	textInputValue := ""
	if !useGrobulkAppend {
		// assign pre-existing grocery list to the textInputValue so that it can be prefilled
		groceryEntries, err := c.groceryEntryRepository.FindByQueryWithConfig(
			&models.GroceryEntry{
				GuildID: guildID,
			},
			repositories.GroceryEntryQueryOpts{
				IsStrongNilForGroceryListID: true,
			},
		)
		if err != nil {
			return nil, err
		}
		textInputValue = getGrobulkInputFromEntries(groceryEntries)
	} else {
		// do nothing here since we will be adding to the existing grocery list, so no prefilled value needed
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "bulk_edit",
					Label:       "Input multiple groceries",
					Placeholder: "Paste your grocery entries here - each entries are separated by newlines.\n\nFor example: \nTea\nBeef",
					Required:    true,
					MaxLength:   4000,
					Style:       discordgo.TextInputParagraph,
					Value:       textInputValue,
				},
			},
		},
	}
	data := &discordgo.InteractionResponseData{
		CustomID:   "grobulk",
		Title:      "!grobulk - add multiple groceries",
		Components: components,
	}
	return data, nil
}
