package slash

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrMissingSlashCommandOption            = errors.New("Cannot find slash command option.")
	ErrMissingOptionKeyForDefaultMarshaller = errors.New("Cannot find mainInputOptionKey, which is required if the default marshaller is used.")
)

type argStrMarshaller = func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error)

type slashCommandHandlerMetadata struct {
	// maintains compatibility with the legacy command handlers
	customArgStrMarshaller argStrMarshaller
	// determines which option to extract the argStr from
	mainInputOptionKey string
	// maybe you want /grolist-new to execute !grolist new instead?
	commandMappingOverride string
}

var (
	defaultListLabelOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "list-label",
		Description: "Label for your custom grocery list",
		Required:    false,
	}
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "gro",
			Description: "Add a grocery entry",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "entry",
					Description: "New grocery entry",
					Required:    true,
				},
				defaultListLabelOption,
			},
		},
		{
			Name:        "groclear",
			Description: "Clears your grocery list",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				defaultListLabelOption,
			},
		},
		{
			Name:        "groremove",
			Description: "Removes a grocery entry",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "entry",
					Description: "The grocery entry (or entry #) to be removed.",
					Required:    true,
				},
				defaultListLabelOption,
			},
		},
		{
			Name:        "groedit",
			Description: "Edits a grocery entry",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "entry-index",
					Description: "The entry # to be edited.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "new-name",
					Description: "The name to edit the entry # to.",
					Required:    true,
				},
				defaultListLabelOption,
			},
		},
		{
			Name:        "grohelp",
			Description: "Get help!",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "grolist",
			Description: "View your current grocery list.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				defaultListLabelOption,
			},
		},
		{
			Name:        "grolist-new",
			Description: "Creates a new grocery list.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "label",
					Description: "The label to put on your new grocery list - will be used to refer to your grocery list.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "pretty-name",
					Description: "A pretty name for your new grocery list - will be used when displaying your grocery list.",
					Required:    false,
				},
			},
		},
		{
			Name:        "grolist-delete",
			Description: "Deletes your grocery list (does not delete your entries - use /groclear for that).",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "list-label",
					Description: "Label for your custom grocery list",
					Required:    true,
				},
			},
		},
	}
	commandsMetadata = map[string]slashCommandHandlerMetadata{
		"gro": {
			mainInputOptionKey: "entry",
		},
		"groremove": {
			mainInputOptionKey: "entry",
		},
		"groedit": {
			customArgStrMarshaller: func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error) {
				entryIndex := int64(-1)
				newName := ""
				for _, o := range options {
					switch o.Name {
					case "entry-index":
						entryIndex = o.IntValue()
					case "new-name":
						newName = o.StringValue()
					}
				}
				if entryIndex == -1 || newName == "" {
					return "", ErrMissingSlashCommandOption
				}
				return fmt.Sprintf("%d %s", entryIndex, newName), nil
			},
		},
		"grolist-new": {
			commandMappingOverride: "!grolist",
			customArgStrMarshaller: func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error) {
				label := ""
				prettyName := ""
				for _, o := range options {
					switch o.Name {
					case "label":
						label = o.StringValue()
					case "pretty-name":
						prettyName = o.StringValue()
					}
				}
				if label == "" {
					return "", ErrMissingSlashCommandOption
				}
				argStr = "new " + label
				if prettyName != "" {
					argStr += " " + prettyName
				}
				return argStr, nil
			},
		},
		"grolist-delete": {
			commandMappingOverride: "!grolist",
			customArgStrMarshaller: func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error) {
				return "delete", nil
			},
		},
	}
)

var defaultSlashCommandArgStrMarshaller argStrMarshaller = func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error) {
	if commandMetadata.mainInputOptionKey == "" {
		return "", ErrMissingOptionKeyForDefaultMarshaller
	}
	for _, option := range options {
		if option.Name == commandMetadata.mainInputOptionKey {
			return option.StringValue(), nil
		}
	}
	return "", ErrMissingSlashCommandOption
}

func getListLabelFromOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (listLabel string) {
	for _, option := range options {
		if option.Name == defaultListLabelOption.Name {
			return option.StringValue()
		}
	}
	return ""
}

func onHandlingErrorRespond(logger *zap.Logger, sess *discordgo.Session, interaction *discordgo.Interaction) {
	if err := sess.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Whoops, I seem to not know how to handle that command. Give it a day and my hooman will fix it!",
		},
	}); err != nil {
		logger.Error("Cannot respond to interaction.", zap.Any("Error", err))
	}
}

func Register(sess *discordgo.Session, db *gorm.DB, logger *zap.Logger, grobotVersion string) (err error, cleanup func(useAllCommands bool) error) {
	targetGuildID := "815482602278354944"
	// TODO do cleanup
	createdCommands, err := sess.ApplicationCommandBulkOverwrite(sess.State.User.ID, targetGuildID, commands)
	if err != nil {
		return err, nil
	}
	cleanupHandler := sess.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer handlers.Recover(logger)
		commandData := i.ApplicationCommandData()
		command := "!" + commandData.Name
		logger.Info("Received slash command.", zap.String("Command", commandData.Name), zap.Any("commandData", commandData))
		listLabel := getListLabelFromOptions(commandData.Options)
		argStrMarshaller := defaultSlashCommandArgStrMarshaller
		commandMetadata, ok := commandsMetadata[commandData.Name]
		argStr := ""
		if ok {
			if commandMetadata.customArgStrMarshaller != nil {
				argStrMarshaller = commandMetadata.customArgStrMarshaller
			}
			marshalledArgStr, err := argStrMarshaller(commandData.Options, &commandMetadata)
			if err != nil {
				onHandlingErrorRespond(logger, sess, i.Interaction)
				return
			}
			argStr = marshalledArgStr
			if commandMetadata.commandMappingOverride != "" {
				command = commandMetadata.commandMappingOverride
			}
		}
		handler := handlers.NewHandler(sess, &handlers.CommandContext{
			Command:           command,
			GrocerySublist:    listLabel,
			ArgStr:            argStr,
			GuildID:           i.GuildID,
			ChannelID:         i.ChannelID,
			CommandSourceType: handlers.CommandSourceSlashCommand,
			Interaction:       i.Interaction,
			AuthorID:          i.Member.User.ID,
		}, db, grobotVersion, logger)
		if err := handler.Handle(); err != nil {
			logger.Error("TEMP: Unable to handle.", zap.Any("Error", err))
			sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Whoops, something went wrong! My hoomans will fix that ASAP (usually within 24h) - sorry for the inconvenience!",
				},
			})
		}
		// s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 	Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		// 	Data: &discordgo.InteractionResponseData{
		// 		Content: "Hey there! Congratulations, you just executed your first slash command",
		// 	},
		// })
	})
	return nil, func(useAllCommands bool) error {
		commands := createdCommands
		if useAllCommands {
			allCommands, err := sess.ApplicationCommands(sess.State.User.ID, targetGuildID)
			if err != nil {
				return err
			}
			commands = allCommands
		}
		for _, cmd := range commands {
			err := sess.ApplicationCommandDelete(sess.State.User.ID, targetGuildID, cmd.ID)
			if err != nil {
				return err
			}
		}
		cleanupHandler()
		return nil
	}
}

func Cleanup(sess *discordgo.Session) error {
	targetGuildID := "815482602278354944"
	commands, err := sess.ApplicationCommands(sess.State.User.ID, targetGuildID)
	if err != nil {
		return err
	}
	for _, cmd := range commands {
		err := sess.ApplicationCommandDelete(sess.State.User.ID, targetGuildID, cmd.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
