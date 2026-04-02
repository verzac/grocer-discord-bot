package modal

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/handlers/slash/defaults"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
)

const checkboxGroupMaxOptions = 10

// Discord counts string length like JS (UTF-16 code units); labels must be 1–100.
const discordCheckboxOptionLabelMaxUTF16 = 100

const groremoveEntryFYIText = "FYI: You do not need to type in the items you'd like to remove manually; you can just select the checkboxes next time."

func utf16UnitsForRune(r rune) int {
	if r > 0xffff {
		return 2
	}
	return 1
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += utf16UnitsForRune(r)
	}
	return n
}

// truncateDiscordLabel returns a prefix of s whose UTF-16 length is at most maxUTF16.
func truncateDiscordLabel(s string, maxUTF16 int) string {
	if maxUTF16 <= 0 {
		return ""
	}
	rs := []rune(s)
	n := 0
	i := 0
	for i < len(rs) {
		add := utf16UnitsForRune(rs[i])
		if n+add > maxUTF16 {
			break
		}
		n += add
		i++
	}
	return string(rs[:i])
}

func groremoveCheckboxOptionLabel(index int, itemDesc string) string {
	prefix := fmt.Sprintf("%d. ", index)
	prefixLen := utf16Len(prefix)
	if prefixLen >= discordCheckboxOptionLabelMaxUTF16 {
		return truncateDiscordLabel(prefix, discordCheckboxOptionLabelMaxUTF16)
	}
	remaining := discordCheckboxOptionLabelMaxUTF16 - prefixLen
	return prefix + truncateDiscordLabel(itemDesc, remaining)
}

func buildGroremoveModalComponents(groceries []models.GroceryEntry, preselected []string) []discordgo.MessageComponent {
	nonce := rand.Int63()
	preselectedSet := make(map[string]struct{}, len(preselected))
	for _, v := range preselected {
		preselectedSet[v] = struct{}{}
	}
	var components []discordgo.MessageComponent
	for chunkStart := 0; chunkStart < len(groceries); chunkStart += checkboxGroupMaxOptions {
		chunkEnd := min(chunkStart+checkboxGroupMaxOptions, len(groceries))
		chunk := groceries[chunkStart:chunkEnd]
		options := make([]discordgo.CheckboxGroupOption, 0, len(chunk))
		for j, g := range chunk {
			absIdx := chunkStart + j
			// absIdx+1 is the 1-based index that OnRemove() expects
			value := strconv.Itoa(absIdx + 1)
			def := false
			opt := discordgo.CheckboxGroupOption{
				Label:   groremoveCheckboxOptionLabel(absIdx+1, g.ItemDesc),
				Value:   value,
				Default: &def,
			}
			if _, ok := preselectedSet[value]; ok {
				def = true
				opt.Default = &def
			}
			options = append(options, opt)
		}
		groupNum := chunkStart / checkboxGroupMaxOptions
		required := false
		components = append(components, discordgo.Label{
			Label: fmt.Sprintf("Items %d-%d", chunkStart+1, chunkEnd),
			Component: discordgo.CheckboxGroup{
				CustomID: fmt.Sprintf("groremove_items_%d_%d", groupNum, nonce),
				Options:  options,
				Required: &required,
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

func getGroremoveCommandContext(i *discordgo.InteractionCreate, commandName string, groceryListRepo repositories.GroceryListRepository) (*handlers.CommandContext, error) {
	data := i.ModalSubmitData()
	selectedIndexes := collectSelectedIndexes(data.Components)
	grocerySublist := ""
	if strings.HasPrefix(commandName, "groremove:") {
		idStr := strings.TrimPrefix(commandName, "groremove:")
		listID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return nil, err
		}
		gl, err := groceryListRepo.GetByQuery(&models.GroceryList{ID: uint(listID), GuildID: i.GuildID})
		if err != nil {
			return nil, err
		}
		if gl == nil {
			return nil, fmt.Errorf("grocery list not found for modal custom_id")
		}
		grocerySublist = gl.ListLabel
	}
	return &handlers.CommandContext{
		Command:                     handlers.CmdGroRemove,
		GrocerySublist:              grocerySublist,
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
	commandData := c.interaction.ApplicationCommandData()
	listLabel := defaults.ListLabelFromSlashOptions(commandData.Options)

	query := &models.GroceryEntry{GuildID: c.guildID}
	queryOpts := repositories.GroceryEntryQueryOpts{}
	modalCustomID := "groremove"
	var groceryList *models.GroceryList

	if listLabel != "" {
		gl, err := c.groceryListRepository.GetByQuery(&models.GroceryList{GuildID: c.guildID, ListLabel: listLabel})
		if err != nil {
			return nil, err
		}
		if gl == nil {
			return nil, c.RespondWithMessageInsteadOfModal(fmt.Sprintf("Whoops, I can't seem to find the grocery list labeled as *%s*.", listLabel))
		}
		groceryList = gl
		query.GroceryListID = &gl.ID
		modalCustomID = fmt.Sprintf("groremove:%d", gl.ID)
	} else {
		queryOpts.IsStrongNilForGroceryListID = true
	}

	groceries, err := c.groceryEntryRepository.FindByQueryWithConfig(query, queryOpts)
	if err != nil {
		return nil, err
	}
	if len(groceries) == 0 {
		return nil, c.RespondWithMessageInsteadOfModal(fmt.Sprintf("Whoops, you do not have any items in %s.", groceryList.GetName()))
	}

	entry := defaults.EntryFromSlashOptions(commandData.Options)
	var preselected []string
	if entry != "" {
		preselected, err = handlers.PreselectedGroremoveOptionValues(entry, groceries, groceryList)
		if err != nil {
			return nil, c.RespondWithMessageInsteadOfModal(fmt.Sprintf(":exploding_head: Oops, something went wrong: %s", err.Error()))
		}
	}

	components := buildGroremoveModalComponents(groceries, preselected)
	if entry != "" {
		components = append([]discordgo.MessageComponent{discordgo.TextDisplay{Content: groremoveEntryFYIText}}, components...)
	}

	return &discordgo.InteractionResponseData{
		CustomID:   modalCustomID,
		Title:      "Remove Groceries",
		Components: components,
	}, nil
}
