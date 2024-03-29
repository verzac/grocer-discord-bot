package slash

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/config"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/handlers/slash/native"
	"github.com/verzac/grocer-discord-bot/monitoring"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

var (
	ErrMissingSlashCommandOption            = errors.New("cannot find slash command option")
	ErrMissingOptionKeyForDefaultMarshaller = errors.New("cannot find mainInputOptionKey, which is required if the default marshaller is used")
	ErrIncorrectFormatInt                   = errors.New("expected a number as an input")
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
		Type:         discordgo.ApplicationCommandOptionString,
		Name:         "list-label",
		Description:  "Label for your custom grocery list.",
		Required:     false,
		Autocomplete: true,
	}
	defaultAllListOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "all",
		Description: "Display all of your grocery lists instead of the one denoted in list-label.",
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
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "entry",
					Description:  "The grocery entry (or entry #) to be removed. Remove multiple items by inputting multiple #.",
					Required:     true,
					Autocomplete: true,
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
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "entry-index",
					Description:  "The entry # to be edited.",
					Required:     true,
					Autocomplete: true,
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
				defaultAllListOption,
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
					Type:         defaultListLabelOption.Type,
					Name:         defaultListLabelOption.Name,
					Description:  defaultListLabelOption.Description,
					Autocomplete: defaultListLabelOption.Autocomplete,
					Required:     true,
				},
			},
		},
		{
			Name:        "groreset",
			Description: "Clear all of your data from GroceryBot.",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "grobulk",
			Description: "Add multiple grocery entries to your list.",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "grohere",
			Description: "Attach a self-updating list for your grocery list to the current channel.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				defaultAllListOption,
				defaultListLabelOption,
			},
		},
		{
			Name:        "gropatron",
			Description: "Do stuff for your account here",
			// Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "register",
					Description: "Registers your Patreon entitlement & benefits for this server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "deregister",
					Description: "Remove your Patreon benefits for this server, so that you can use it for other servers.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
		{
			Name:        "developer",
			Description: "Create a new API Client ID & Secret so that you can integrate directly with GroceryBot!",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "config",
			Description: "Customize GroceryBot for your server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "set",
					Description: "Configure GroceryBot for your server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "use_ephemeral",
							Description: "Enable ephemeral message replies from GroceryBot, which are only visible to you and will disappear.",
							Required:    true,
						},
					},
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
						s, err := strconv.Atoi(o.StringValue())
						if err != nil {
							return "", ErrIncorrectFormatInt
						}
						entryIndex = int64(s)
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
		"grolist": {
			customArgStrMarshaller: func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error) {
				for _, o := range options {
					if o.Name == defaultAllListOption.Name {
						return "all", nil
					}
				}
				return "", nil
			},
		},
		"grohere": {
			customArgStrMarshaller: func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error) {
				for _, o := range options {
					if o.Name == defaultAllListOption.Name {
						return "all", nil
					}
				}
				return "", nil
			},
		},
		"gropatron": {
			customArgStrMarshaller: func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error) {
				for _, o := range options {
					if o.Type == discordgo.ApplicationCommandOptionSubCommand {
						argStr += o.Name
					}
				}
				return argStr, nil
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

func isSkippableError(err error) bool {
	if discordErr, ok := err.(*discordgo.RESTError); ok && discordErr.Response.StatusCode == 403 {
		// if 403, assume that they have no access to the guild - continue anyways
		return true
	}
	return false
}

func getCommandName(i *discordgo.InteractionCreate) string {
	switch i.Type {
	case discordgo.InteractionModalSubmit:
		return i.ModalSubmitData().CustomID
	case discordgo.InteractionMessageComponent:
		return i.MessageComponentData().CustomID
	default:
		return i.ApplicationCommandData().Name
	}
}

func getMessageCommandContext(i *discordgo.InteractionCreate, commandName string) (*handlers.CommandContext, error) {
	command := "!" + commandName
	commandData := i.ApplicationCommandData()
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
			return nil, err
		}
		argStr = marshalledArgStr
		if commandMetadata.commandMappingOverride != "" {
			command = commandMetadata.commandMappingOverride
		}
	}
	commandContext := &handlers.CommandContext{
		Command:                     command,
		GrocerySublist:              listLabel,
		ArgStr:                      argStr,
		GuildID:                     i.GuildID,
		ChannelID:                   i.ChannelID,
		CommandSourceType:           handlers.CommandSourceSlashCommand,
		Interaction:                 i.Interaction,
		AuthorID:                    i.Member.User.ID,
		AuthorUsername:              i.Member.User.Username,
		AuthorUsernameDiscriminator: i.Member.User.Discriminator,
	}
	return commandContext, nil
}

func getCommandContext(i *discordgo.InteractionCreate, commandName string) (*handlers.CommandContext, error) {
	switch i.Type {
	case discordgo.InteractionModalSubmit:
		return getCommandContextFromModalSubmission(i, commandName)
	default:
		return getMessageCommandContext(i, commandName)
	}
}

func registerForGuild(sess *discordgo.Session, logger *zap.Logger, targetGuildID string, commands []*discordgo.ApplicationCommand) ([]*discordgo.ApplicationCommand, error) {
	loopLog := logger.With(zap.String("TargetGuildID", targetGuildID)).Named("registration")
	createdCommands, err := sess.ApplicationCommandBulkOverwrite(sess.State.User.ID, targetGuildID, commands)
	if err != nil {
		if isSkippableError(err) {
			// if 403, assume that they have no access to the guild - continue anyways
			loopLog.Info(fmt.Sprintf("Skipping slash command registration for %s.", targetGuildID), zap.Any("Error", err))
			return nil, nil
		}
		return nil, err
	}
	return createdCommands, nil
}

func Register(sess *discordgo.Session, db *gorm.DB, logger *zap.Logger, grobotVersion string, cw *cloudwatch.CloudWatch) (cleanup func(useAllCommands bool) error, err error) {
	createdCommandsMap := make(map[string][]*discordgo.ApplicationCommand, 0)
	logger = logger.Named("registration")
	targetGuildIDs := config.GetGuildIDsToRegisterSlashCommandsOn()
	logger.Info("Starting the registration process for guilds.", zap.Any("targetGuildIDs", targetGuildIDs))
	ignoredSlashCommands := config.GetIgnoredSlashCommands(grobotVersion)
	if len(ignoredSlashCommands) > 0 {
		logger.Info("Ignoring slash commands.", zap.Any("ignoredSlashCommands", ignoredSlashCommands))
	} else {
		logger.Debug("No slash commands to ignore.")
	}
	whitelistedSlashCommands := config.GetWhitelistedSlashCommands(grobotVersion)
	commandsToRegister := make([]*discordgo.ApplicationCommand, 0, len(commands))
	commandsToRegisterForGuild := map[string][]*discordgo.ApplicationCommand{}
	for _, cmd := range commands {
		if _, ok := ignoredSlashCommands[cmd.Name]; !ok {
			commandsToRegister = append(commandsToRegister, cmd)
		}
		for guildID, commandSet := range whitelistedSlashCommands {
			if _, ok := commandSet[cmd.Name]; ok {
				// register the command for guild
				if _, ok := commandsToRegisterForGuild[guildID]; !ok {
					commandsToRegisterForGuild[guildID] = []*discordgo.ApplicationCommand{}
				}
				commandsToRegisterForGuild[guildID] = append(commandsToRegisterForGuild[guildID], cmd)
			}
		}
	}
	for _, targetGuildID := range targetGuildIDs {
		createdCommands, err := registerForGuild(sess, logger, targetGuildID, commandsToRegister)
		if err != nil {
			return nil, err
		} else if createdCommands != nil {
			createdCommandsMap[targetGuildID] = createdCommands
		}
	}
	if len(commandsToRegisterForGuild) > 0 {
		logger.Info("Registering guild-specific commands...", zap.Any("commandsToRegisterForGuild", commandsToRegisterForGuild))
	} else {
		logger.Debug("No guild-specific commands detected.")
	}
	for targetGuildID, commandsToRegister := range commandsToRegisterForGuild {
		createdCommands, err := registerForGuild(sess, logger, targetGuildID, commandsToRegister)
		if err != nil {
			return nil, err
		} else if createdCommands != nil {
			createdCommandsMap[targetGuildID] = createdCommands
		}
	}
	cleanupHandler := sess.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer handlers.Recover(logger)
		if i.Member == nil {
			if err := sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "I do not currently support commands through my DM. Please put me in a server and I'll work magic for you!",
				},
			}); err != nil {
				logger.Error("Cannot respond to DM interaction.", zap.Error(err))
			}
			return
		}
		if modalCreationCtx := NewModalCreationContext(sess, db, logger, i, grobotVersion); modalCreationCtx != nil {
			modalCreationCtx.Handle()
			return
		}
		commandName := getCommandName(i)
		logger := logger.Named("handler").With(
			zap.String("GuildID", i.GuildID),
			zap.String("SlashCommandName", commandName),
		)
		if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			err := NewAutoCompleteHandler(sess, db, logger, i).Handle()
			if err != nil {
				logger.Error("Failed to handle auto-complete event.", zap.Error(err))
			}
			return
		}
		// proxy to message handler
		command := "!" + commandName
		cm := monitoring.NewCommandMetric(cw, command, logger)
		defer cm.Done()
		logger.Debug("Received slash command.")
		if isHandled := native.Handle(native.NativeSlashHandlingParams{
			Session:           s,
			InteractionCreate: i,
			CommandName:       commandName,
			DB:                db,
			Logger:            logger,
		}); isHandled {
			return
		}
		commandContext, err := getCommandContext(i, commandName)
		if err != nil {
			onHandlingErrorRespond(logger, sess, i.Interaction)
			return
		}
		handler := handlers.NewHandler(sess, commandContext, db, grobotVersion, logger)
		if err := handler.Handle(); err != nil {
			logger.Error("Unable to handle.", zap.Any("Error", err))
			if err := sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Whoops, something went wrong! My hoomans will fix that ASAP (usually within 24h) - sorry for the inconvenience!",
				},
			}); err != nil {
				handler.LogError(err)
			}
		}
	})
	return func(useAllCommands bool) error {
		for guildID, cmds := range createdCommandsMap {
			for _, cmd := range cmds {
				logger.Info("Deleting command.", zap.String("CommandName", cmd.Name))
				err := sess.ApplicationCommandDelete(sess.State.User.ID, guildID, cmd.ID)
				if err != nil {
					return err
				}
			}
		}
		cleanupHandler()
		return nil
	}, nil
}

func CleanupCommands(sess *discordgo.Session, logger *zap.Logger, targetGuildID string, cmds []*discordgo.ApplicationCommand) []error {
	errChan := make(chan error, len(cmds))
	cmdChan := make(chan *discordgo.ApplicationCommand)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			for cmd := range cmdChan {
				l := logger.With(zap.String("CommandName", cmd.Name), zap.Int("GoroutineID", goroutineID))
				err := sess.ApplicationCommandDelete(sess.State.User.ID, targetGuildID, cmd.ID)
				if err != nil {
					errChan <- err
				}
				l.Debug("Application command deleted.")
			}
			wg.Done()
		}(i)
	}
	for _, cmd := range cmds {
		cmdChan <- cmd
	}
	close(cmdChan)
	wg.Wait()
	close(errChan)
	errs := make([]error, 0, len(cmds))
	for err := range errChan {
		errs = append(errs, err)
	}
	return errs
}

func Cleanup(sess *discordgo.Session, logger *zap.Logger) error {
	logger = logger.Named("manualcleanup")
	guildIDsToCleanup := config.GetGuildIDsToDeregisterSlashCommandsFrom()
	logger.Info("Starting cleanup process.", zap.Any("guildIDsToCleanup", guildIDsToCleanup))
	for _, targetGuildID := range guildIDsToCleanup {
		loopLog := logger.With(zap.String("TargetGuildID", targetGuildID))
		loopLog.Debug("Retrieving all commands.")
		registeredCommands, err := sess.ApplicationCommands(sess.State.User.ID, targetGuildID)
		if err != nil {
			if isSkippableError(err) {
				// if 403, assume that they have no access to the guild - continue anyways
				loopLog.Info(fmt.Sprintf("Skipping slash command cleanup for %s.", targetGuildID), zap.Any("Error", err))
				continue
			}
			return err
		}
		loopLog.Debug("All commands retrieved.",
			zap.Array("Commands", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
				for _, c := range registeredCommands {
					ae.AppendString(c.Name)
				}
				return nil
			})),
		)
		if errs := CleanupCommands(sess, loopLog, targetGuildID, registeredCommands); len(errs) > 0 {
			loopLog.Error("Encountered error while deleting commands.", zap.Any("Errors", errs))
			errMsg := ""
			for _, err := range errs {
				errMsg += err.Error()
			}
			return errors.New(errMsg)
		}
	}
	logger.Info("Commands have been cleaned up successfully.")
	return nil
}
