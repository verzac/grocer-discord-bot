package slash

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrMissingSlashCommandOption = errors.New("Cannot find slash command option.")
)

type argStrMarshaller = func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error)

type slashCommandHandlerMetadata struct {
	// maintains compatibility with
	customArgStrMarshaller argStrMarshaller
	mainInputOptionKey     string
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
	}
	commandsMetadata = map[string]slashCommandHandlerMetadata{
		"gro": {
			mainInputOptionKey: "entry",
		},
	}
)

var defaultSlashCommandArgStrMarshaller argStrMarshaller = func(options []*discordgo.ApplicationCommandInteractionDataOption, commandMetadata *slashCommandHandlerMetadata) (argStr string, err error) {
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
		logger.Info("Received slash command.", zap.String("Command", command), zap.Any("commandData", commandData))
		listLabel := getListLabelFromOptions(commandData.Options)
		argStrMarshaller := defaultSlashCommandArgStrMarshaller
		commandMetadata, ok := commandsMetadata[commandData.Name]
		if !ok {
			onHandlingErrorRespond(logger, sess, i.Interaction)
			return
		}
		if commandMetadata.customArgStrMarshaller != nil {
			argStrMarshaller = commandMetadata.customArgStrMarshaller
		}
		argStr, err := argStrMarshaller(commandData.Options, &commandMetadata)
		if err != nil {
			onHandlingErrorRespond(logger, sess, i.Interaction)
			return
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
