package slash

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ModalCreationContext struct {
	groceryListRepository repositories.GroceryListRepository
	session               *discordgo.Session
	logger                *zap.Logger
	interaction           *discordgo.InteractionCreate
	commandName           string
	guildID               string
	authorID              string
	grobotVersion         string
}

const defaultGroceryListValue = ":"

var (
	ErrCannotFindMatchingCommandContextGetter = errors.New("cannot find matching command context getter")
)

var (
	modalCommandHandlers = map[string]func(c *ModalCreationContext) (*discordgo.InteractionResponseData, error){
		"grobulk": func(c *ModalCreationContext) (*discordgo.InteractionResponseData, error) {
			lists, err := c.groceryListRepository.FindByQuery(&models.GroceryList{
				GuildID: c.guildID,
			})
			if err != nil {
				return nil, err
			}
			groceryListSelectOptions := []discordgo.SelectMenuOption{}
			groceryListSelectOptions = append(groceryListSelectOptions, discordgo.SelectMenuOption{
				Label:   "Default List",
				Value:   defaultGroceryListValue,
				Default: true,
			})
			for _, l := range lists {
				groceryListSelectOptions = append(groceryListSelectOptions, discordgo.SelectMenuOption{
					Label: l.GetName(),
					Value: l.ListLabel,
				})
			}
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "bulk_input",
							Label:       "Input multiple groceries",
							Placeholder: "Paste your grocery entries here (each entries are separated by newlines)!",
							Required:    true,
							MaxLength:   1024,
							Style:       discordgo.TextInputParagraph,
						},
					},
				},
			}
			if len(groceryListSelectOptions) > 0 {
				minValues := 0
				components = append(components, discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:  "Which grocery list do you want to put it into?",
							MinValues: &minValues,
							MaxValues: 1,
							Options:   groceryListSelectOptions,
						},
					},
				})
			}
			data := &discordgo.InteractionResponseData{
				CustomID:   "grobulk",
				Title:      "!grobulk - add multiple groceries",
				Components: components,
			}
			return data, nil
		},
	}
)

func NewModalCreationContext(sess *discordgo.Session, db *gorm.DB, logger *zap.Logger, i *discordgo.InteractionCreate, grobotVersion string) (c *ModalCreationContext) {
	commandKey := ""
	isHandleable := false
	switch i.Type {
	// case discordgo.InteractionModalSubmit:
	// 	isHandleable = true
	// 	commandKey = i.ModalSubmitData().CustomID
	case discordgo.InteractionApplicationCommand:
		commandKey = i.ApplicationCommandData().Name
		for k := range modalCommandHandlers {
			if k == commandKey {
				isHandleable = true
				break
			}
		}
	}
	if !isHandleable {
		return nil
	}
	return &ModalCreationContext{
		groceryListRepository: &repositories.GroceryListRepositoryImpl{DB: db},
		commandName:           commandKey,
		guildID:               i.GuildID,
		authorID:              i.Member.User.ID, // check nils in caller
		session:               sess,
		logger:                logger.Named("modal"),
		interaction:           i,
		grobotVersion:         grobotVersion,
	}
}

func (c *ModalCreationContext) Handle() {
	defer handlers.Recover(c.logger)
	c.logger.Debug("Creating modal.", zap.Uint8("InteractionType", uint8(c.interaction.Type)))
	switch c.interaction.Type {
	default:
		if handler, ok := modalCommandHandlers[c.commandName]; ok {
			data, err := handler(c)
			if err != nil {
				c.logger.Error("Unable to handle command.", zap.Error(err))
				return
			}
			if err := c.session.InteractionRespond(c.interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: data,
			}); err != nil {
				c.logger.Error("Unable to respond to interaction.", zap.Error(err))
			}
		} else {
			c.logger.Error("Unable to find modal command handler. Did we check correctly before running .Handle()?")
			return
		}
	}
}

var (
	commandContextGetters = map[string]func(i *discordgo.InteractionCreate, commandName string) (*handlers.CommandContext, error){
		"grobulk": func(i *discordgo.InteractionCreate, commandName string) (*handlers.CommandContext, error) {
			return &handlers.CommandContext{
				Command:                     "!" + commandName,
				GrocerySublist:              "",
				ArgStr:                      "chicken", // TODO ASSIGN ME
				GuildID:                     i.GuildID,
				AuthorID:                    i.Member.User.ID, // nil-check this in caller
				ChannelID:                   i.ChannelID,
				AuthorUsername:              i.Member.User.Username,
				AuthorUsernameDiscriminator: i.Member.User.Discriminator,
				CommandSourceType:           handlers.CommandSourceSlashCommand,
				Interaction:                 i.Interaction,
			}, nil
		},
	}
)

func getCommandContextFromModalSubmission(i *discordgo.InteractionCreate, commandName string) (*handlers.CommandContext, error) {
	if getter, ok := commandContextGetters[commandName]; ok {
		return getter(i, commandName)
	} else {
		return nil, ErrCannotFindMatchingCommandContextGetter
	}
}
