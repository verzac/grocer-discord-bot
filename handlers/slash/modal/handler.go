package modal

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrCannotFindMatchingCommandContextGetter = errors.New("cannot find matching command context getter")
	ErrAlreadyHandled                         = errors.New("interaction already handled")
)

var (
	modalCommandHandlers = map[string]func(c *ModalCreationContext) (*discordgo.InteractionResponseData, error){
		"grobulk":   handleGrobulkCommand,
		"groremove": handleGroremoveCommand,
	}
	commandContextGetters = map[string]func(i *discordgo.InteractionCreate, commandName string, groceryListRepo repositories.GroceryListRepository) (*handlers.CommandContext, error){
		"grobulk":   getGrobulkCommandContext,
		"groremove": getGroremoveCommandContext,
	}
)

func modalSubmitGetterKey(customID string) string {
	if customID == "groremove" || strings.HasPrefix(customID, "groremove:") {
		return "groremove"
	}
	return customID
}

func GetCommandContextFromModalSubmission(i *discordgo.InteractionCreate, commandName string, groceryListRepo repositories.GroceryListRepository) (*handlers.CommandContext, error) {
	key := modalSubmitGetterKey(commandName)
	if getter, ok := commandContextGetters[key]; ok {
		return getter(i, commandName, groceryListRepo)
	}
	return nil, ErrCannotFindMatchingCommandContextGetter
}

type ModalCreationContext struct {
	groceryListRepository  repositories.GroceryListRepository
	groceryEntryRepository repositories.GroceryEntryRepository
	guildConfigRepository  repositories.GuildConfigRepository
	session                *discordgo.Session
	logger                 *zap.Logger
	interaction            *discordgo.InteractionCreate
	cachedConfig           *models.GuildConfig
	commandName            string
	guildID                string
	authorID               string
	grobotVersion          string
}

func (c *ModalCreationContext) Handle() {
	defer handlers.Recover(c.logger)
	c.logger.Debug("Creating modal.", zap.Uint8("InteractionType", uint8(c.interaction.Type)))
	switch c.interaction.Type {
	default:
		if handler, ok := modalCommandHandlers[c.commandName]; ok {
			data, err := handler(c)
			if err != nil {
				if errors.Is(err, ErrAlreadyHandled) {
					return
				}
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

func (c *ModalCreationContext) getConfig() (*models.GuildConfig, error) {
	if c.cachedConfig != nil {
		return c.cachedConfig, nil
	}
	guildID := c.guildID
	config, err := c.guildConfigRepository.Get(guildID)
	if err != nil {
		return nil, err
	}
	c.cachedConfig = config
	return config, nil
}

func (c *ModalCreationContext) RespondWithMessageInsteadOfModal(msg string) error {
	flags := discordgo.MessageFlags(0)
	config, err := c.getConfig()
	if err != nil {
		c.logger.Error("Failed to load guild config. Non-critical error, skipping.", zap.Error(err))
	}
	if config != nil && config.UseEphemeral {
		flags |= discordgo.MessageFlagsEphemeral
	}

	if err := c.session.InteractionRespond(c.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   flags,
		},
	}); err != nil {
		return err
	}
	return ErrAlreadyHandled
}

func NewModalCreationContext(sess *discordgo.Session, db *gorm.DB, logger *zap.Logger, i *discordgo.InteractionCreate, grobotVersion string) (c *ModalCreationContext) {
	commandKey := ""
	isHandleable := false
	switch i.Type {
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
		groceryListRepository:  &repositories.GroceryListRepositoryImpl{DB: db},
		groceryEntryRepository: &repositories.GroceryEntryRepositoryImpl{DB: db},
		guildConfigRepository:  &repositories.GuildConfigRepositoryImpl{DB: db},
		commandName:            commandKey,
		guildID:                i.GuildID,
		authorID:               i.Member.User.ID, // check nils in caller
		session:                sess,
		logger:                 logger.Named("modal"),
		interaction:            i,
		grobotVersion:          grobotVersion,
	}
}
