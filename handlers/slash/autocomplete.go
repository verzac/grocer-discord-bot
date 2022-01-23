package slash

import (
	"errors"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrAutocompleteCommandNotRecognised = errors.New("Not sure which command this auto-complete event is for...")
	ErrAutocompleteMissingOption        = errors.New("Missing option.")
)

type AutocompleteHandler struct {
	logger           *zap.Logger
	groceryEntryRepo repositories.GroceryEntryRepository
	groceryListRepo  repositories.GroceryListRepository
	interaction      *discordgo.InteractionCreate
	sess             *discordgo.Session
	guildID          string
	commandData      *discordgo.ApplicationCommandInteractionData
	nameToOptionsMap map[string]*discordgo.ApplicationCommandInteractionDataOption
}

func NewAutoCompleteHandler(sess *discordgo.Session, db *gorm.DB, logger *zap.Logger, interaction *discordgo.InteractionCreate) *AutocompleteHandler {
	commandData := interaction.ApplicationCommandData()
	nameToOptionsMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, o := range interaction.ApplicationCommandData().Options {
		nameToOptionsMap[o.Name] = o
	}
	return &AutocompleteHandler{
		logger: logger.Named("autocomplete"),
		groceryEntryRepo: &repositories.GroceryEntryRepositoryImpl{
			DB: db,
		},
		interaction: interaction,
		guildID:     interaction.GuildID,
		groceryListRepo: &repositories.GroceryListRepositoryImpl{
			DB: db,
		},
		sess:             sess,
		commandData:      &commandData,
		nameToOptionsMap: nameToOptionsMap,
	}
}

func (a *AutocompleteHandler) GetGroceryList() (*models.GroceryList, error) {
	sublistLabel := ""
	if listOption, ok := a.nameToOptionsMap[defaultListLabelOption.Name]; ok {
		sublistLabel = listOption.StringValue()
	}
	if sublistLabel == "" {
		return nil, nil
	}
	return a.groceryListRepo.GetByQuery(&models.GroceryList{
		GuildID:   a.guildID,
		ListLabel: sublistLabel,
	})
}

func (a *AutocompleteHandler) Handle() error {
	defer handlers.Recover(a.logger)
	a.logger.Debug("Handling autocomplete.")
	if listOption, ok := a.nameToOptionsMap[defaultListLabelOption.Name]; ok && listOption.Focused {
		return a.sess.InteractionRespond(a.interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: a.GetGroceryListChoices(listOption.StringValue()),
			},
		})
	}
	switch "!" + a.commandData.Name {
	case handlers.CmdGroRemove:
		entry, ok := a.nameToOptionsMap["entry"]
		if !ok {
			return ErrAutocompleteMissingOption
		}
		if entry.Focused == true {
			choices := a.GetGroceryEntryChoices(entry.StringValue())
			return a.sess.InteractionRespond(a.interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: choices,
				},
			})
		}
	case handlers.CmdGroEdit:
		entryIndex, ok := a.nameToOptionsMap["entry-index"]
		if !ok {
			return ErrAutocompleteMissingOption
		}
		if entryIndex.Focused == true {
			choices := a.GetGroceryEntryChoices(entryIndex.StringValue())
			return a.sess.InteractionRespond(a.interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: choices,
				},
			})
		}
	default:
		return ErrAutocompleteCommandNotRecognised
	}
	return nil
}

func (a *AutocompleteHandler) GetGroceryListChoices(queryString string) []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	groceryLists, err := a.groceryListRepo.FindByQuery(&models.GroceryList{GuildID: a.guildID})
	if err != nil {
		a.logger.Error("Failed to load grocery lists.", zap.Error(err))
		return choices
	}
	for _, gl := range groceryLists {
		if strings.Contains(strings.ToLower(gl.ListLabel), strings.ToLower(queryString)) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  gl.GetTitle(),
				Value: gl.ListLabel,
			})
		}
	}
	return choices
}

func (a *AutocompleteHandler) GetGroceryEntryChoices(queryString string) []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	groceryList, err := a.GetGroceryList()
	if err != nil {
		// todo handle err
		a.logger.Error("Failed to load grocery entries.", zap.Error(err))
		return choices
	}
	groceries, err := a.groceryEntryRepo.FindByQueryWithConfig(
		&models.GroceryEntry{
			GuildID:       a.guildID,
			GroceryListID: groceryList.GetID(),
		},
		repositories.GroceryEntryQueryOpts{
			IsStrongNilForGroceryListID: true,
		},
	)
	for idx, g := range groceries {
		if strings.Contains(strings.ToLower(g.ItemDesc), strings.ToLower(queryString)) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  g.ItemDesc,
				Value: strconv.Itoa(idx + 1),
			})
		}
		if len(choices) >= 25 {
			break
		}
	}
	return choices
}
