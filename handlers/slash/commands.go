package slash

import (
	"errors"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
		Description: "Label for your custom grocery list.",
		Required:    false,
	}
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "gro",
			Description: "Add a grocery entry.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "entry",
					Description: "Your new grocery entry.",
					Required:    true,
				},
				defaultListLabelOption,
			},
		},
		{
			Name:        "groclear",
			Description: "Clear your grocery list.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				defaultListLabelOption,
			},
		},
		{
			Name:        "groremove",
			Description: "Remove a grocery entry.",
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
			Description: "Edit a grocery entry.",
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
			Description: "Create a new grocery list.",
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
			Description: "Delete your grocery list (does not delete your entries - use /groclear for that).",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "list-label",
					Description: "Label for your custom grocery list.",
					Required:    true,
				},
			},
		},
		{
			Name:        "groreset",
			Description: "Clear all of your data from GroceryBot.",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "grohere",
			Description: "Attach a self-updating list for your grocery list to the current channel.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "all",
					Description: "Display all of your grocery lists instead of the one denoted in list-label.",
					Required:    false,
				},
				defaultListLabelOption,
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
			logger.Info("Retrieving all commands.")
			allCommands, err := sess.ApplicationCommands(sess.State.User.ID, targetGuildID)
			if err != nil {
				return err
			}
			commands = allCommands
		}
		for _, cmd := range commands {
			logger.Info("Deleting command.", zap.String("CommandName", cmd.Name))
			err := sess.ApplicationCommandDelete(sess.State.User.ID, targetGuildID, cmd.ID)
			if err != nil {
				return err
			}
		}
		cleanupHandler()
		return nil
	}
}

func Cleanup(sess *discordgo.Session, logger *zap.Logger) error {
	targetGuildID := "815482602278354944"
	logger.Info("Starting cleanup process.")
	logger.Info("Retrieving all commands.")
	commands, err := sess.ApplicationCommands(sess.State.User.ID, targetGuildID)
	logger.Info("All commands retrieved.",
		zap.Array("Commands", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
			for _, c := range commands {
				ae.AppendString(c.Name)
			}
			return nil
		})),
	)
	if err != nil {
		return err
	}
	errChan := make(chan error, len(commands))
	cmdChan := make(chan *discordgo.ApplicationCommand)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for cmd := range cmdChan {
				l := logger.With(zap.String("CommandName", cmd.Name))
				err := sess.ApplicationCommandDelete(sess.State.User.ID, targetGuildID, cmd.ID)
				if err != nil {
					errChan <- err
				}
				l.Info("Application command deleted.")
			}
			wg.Done()
		}()
	}
	for _, cmd := range commands {
		cmdChan <- cmd
	}
	close(cmdChan)
	wg.Wait()
	close(errChan)
	logger.Info("De-registration complete.")
	errs := make([]error, 0, len(commands))
	for err := range errChan {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		logger.Error("Encountered error while deleting commands.", zap.Any("Errors", errs))
		errMsg := ""
		for _, err := range errs {
			errMsg += err.Error()
		}
		return errors.New(errMsg)
	}
	logger.Info("Commands have been cleaned up successfully.")
	return nil
}
