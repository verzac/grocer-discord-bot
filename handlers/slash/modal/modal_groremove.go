package modal

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
)

const (
	groremoveListLabelOptionName = "list-label"
	componentTypeLabel           = 18
	componentTypeCheckboxGroup   = 22
	maxGroremoveCheckboxGroups   = 5
	maxGroremoveOptionsPerGroup  = 10
	maxGroremoveSelectableItems  = maxGroremoveCheckboxGroups * maxGroremoveOptionsPerGroup
)

type rawJSONMessageComponent struct {
	jsonBlob []byte
}

func (r rawJSONMessageComponent) MarshalJSON() ([]byte, error) {
	return r.jsonBlob, nil
}

func (r rawJSONMessageComponent) Type() discordgo.ComponentType {
	return discordgo.ComponentType(componentTypeLabel)
}

func groceryListLabelHash(listLabel string) string {
	sum := sha1.Sum([]byte(listLabel))
	return hex.EncodeToString(sum[:])
}

func handleGroremoveCommand(c *ModalCreationContext) (*discordgo.InteractionResponseData, error) {
	guildID := c.guildID
	data := c.interaction.ApplicationCommandData()
	listLabel := ""
	for _, o := range data.Options {
		if o.Name == groremoveListLabelOptionName && o.Type == discordgo.ApplicationCommandOptionString {
			listLabel = o.StringValue()
		}
	}

	var groceryList *models.GroceryList
	if listLabel != "" {
		gl, err := c.groceryListRepository.GetByQuery(&models.GroceryList{
			GuildID:   guildID,
			ListLabel: listLabel,
		})
		if err != nil {
			return nil, err
		}
		if gl == nil {
			return &discordgo.InteractionResponseData{
				CustomID: "groremove",
				Title:    "!groremove",
				Components: []discordgo.MessageComponent{
					discordgo.TextDisplay{
						Content: fmt.Sprintf("Whoops, I can't seem to find a grocery list labeled *%s*.", listLabel),
					},
				},
			}, nil
		}
		groceryList = gl
	}

	var groceryListID *uint
	if groceryList != nil {
		groceryListID = &groceryList.ID
	}
	groceryEntries, err := c.groceryEntryRepository.FindByQueryWithConfig(
		&models.GroceryEntry{
			GuildID:       guildID,
			GroceryListID: groceryListID,
		},
		repositories.GroceryEntryQueryOpts{
			IsStrongNilForGroceryListID: true,
		},
	)
	if err != nil {
		return nil, err
	}

	listName := "your grocery list"
	if groceryList != nil {
		listName = groceryList.GetName()
	}

	if len(groceryEntries) == 0 {
		return &discordgo.InteractionResponseData{
			CustomID: "groremove",
			Title:    "!groremove",
			Components: []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: fmt.Sprintf("You do not have any items in %s.", listName),
				},
			},
		}, nil
	}

	if len(groceryEntries) > maxGroremoveSelectableItems {
		return &discordgo.InteractionResponseData{
			CustomID: "groremove",
			Title:    "!groremove",
			Components: []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: fmt.Sprintf(
						"Your list has more than %d items. Discord only allows up to %d checkboxes in this modal.\n\nPlease use `/grobulk` to edit your list.",
						maxGroremoveSelectableItems,
						maxGroremoveSelectableItems,
					),
				},
			},
		}, nil
	}

	hash := groceryListLabelHash(listLabel)
	components := make([]discordgo.MessageComponent, 0, maxGroremoveCheckboxGroups)
	groupIdx := 0
	for start := 0; start < len(groceryEntries); start += maxGroremoveOptionsPerGroup {
		end := start + maxGroremoveOptionsPerGroup
		if end > len(groceryEntries) {
			end = len(groceryEntries)
		}
		slice := groceryEntries[start:end]
		opts := make([]map[string]interface{}, 0, len(slice))
		for i, ge := range slice {
			idx := start + i + 1
			opts = append(opts, map[string]interface{}{
				"label":   fmt.Sprintf("- %s", ge.ItemDesc),
				"value":   fmt.Sprintf("%d", idx),
				"default": false,
			})
		}
		customID := fmt.Sprintf("groremove:list_hash=%s:group=%d", hash, groupIdx)
		groupJSON, err := json.Marshal(map[string]interface{}{
			"type":       componentTypeCheckboxGroup,
			"custom_id":  customID,
			"options":    opts,
			"disabled":   false,
			"required":   false,
			"min_values": 0,
			"max_values": len(opts),
		})
		if err != nil {
			return nil, err
		}
		labelTitle := fmt.Sprintf("Items %d–%d", start+1, end)
		labelJSON, err := json.Marshal(map[string]interface{}{
			"type":      componentTypeLabel,
			"label":     labelTitle,
			"component": json.RawMessage(groupJSON),
		})
		if err != nil {
			return nil, err
		}
		components = append(components, rawJSONMessageComponent{jsonBlob: labelJSON})
		groupIdx++
	}

	return &discordgo.InteractionResponseData{
		CustomID:   "groremove",
		Title:      fmt.Sprintf("Remove from %s", listName),
		Components: components,
	}, nil
}
