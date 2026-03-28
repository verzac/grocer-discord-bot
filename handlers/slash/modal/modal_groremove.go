package modal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
)

const checkboxGroupMaxOptions = 10

func buildGroremoveModalComponents(groceries []models.GroceryEntry) []discordgo.MessageComponent {
	var components []discordgo.MessageComponent
	for chunkStart := 0; chunkStart < len(groceries); chunkStart += checkboxGroupMaxOptions {
		chunkEnd := min(chunkStart+checkboxGroupMaxOptions, len(groceries))
		chunk := groceries[chunkStart:chunkEnd]
		options := make([]discordgo.RadioGroupOption, 0, len(chunk))
		for j, g := range chunk {
			absIdx := chunkStart + j
			// absIdx+1 is the 1-based index that OnRemove() expects
			options = append(options, discordgo.RadioGroupOption{
				Label: fmt.Sprintf("%d. %s", absIdx+1, g.ItemDesc),
				Value: strconv.Itoa(absIdx + 1),
			})
		}
		groupNum := chunkStart / checkboxGroupMaxOptions
		components = append(components, discordgo.Label{
			Label: fmt.Sprintf("Items %d-%d", chunkStart+1, chunkEnd),
			Component: discordgo.CheckboxGroup{
				CustomID: fmt.Sprintf("groremove_items_%d", groupNum),
				Options:  options,
			},
		})
	}
	return components
}

func collectSelectedIndexes(components []discordgo.MessageComponent) []string {
	var selectedIndexes []string
	for _, comp := range components {
		label, ok := comp.(*discordgo.Label)
		if !ok {
			continue
		}
		checkboxGroup, ok := label.Component.(*discordgo.CheckboxGroup)
		if !ok {
			continue
		}
		selectedIndexes = append(selectedIndexes, checkboxGroup.Values...)
	}
	return selectedIndexes
}

func getGroremoveCommandContext(i *discordgo.InteractionCreate, commandName string) (*handlers.CommandContext, error) {
	data := i.ModalSubmitData()
	selectedIndexes := collectSelectedIndexes(data.Components)
	return &handlers.CommandContext{
		Command:                     handlers.CmdGroRemove,
		GrocerySublist:              "",
		ArgStr:                      strings.Join(selectedIndexes, " "),
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
	groceries, err := c.groceryEntryRepository.FindByQueryWithConfig(
		&models.GroceryEntry{GuildID: c.guildID},
		repositories.GroceryEntryQueryOpts{IsStrongNilForGroceryListID: true},
	)
	if err != nil {
		return nil, err
	}
	if len(groceries) == 0 {
		return nil, c.RespondWithMessageInsteadOfModal("Whoops, you don't have any items in your grocery list!")
	}
	return &discordgo.InteractionResponseData{
		CustomID:   "groremove",
		Title:      "Remove Groceries",
		Components: buildGroremoveModalComponents(groceries),
	}, nil
}
