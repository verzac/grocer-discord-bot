package native

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"github.com/verzac/grocer-discord-bot/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	errMissingSubcommand = errors.New("missing subcommand - possible broken mapping in GroceryBot")
)

type NativeSlashHandlingContext struct {
	s                     *discordgo.Session
	i                     *discordgo.InteractionCreate
	apiClientRepository   repositories.ApiClientRepository
	logger                *zap.Logger
	replyCount            int
	guildConfigRepository repositories.GuildConfigRepository
	cachedConfig          *models.GuildConfig
}

type replyOptions struct {
	IsPrivate bool
}

func (c *NativeSlashHandlingContext) replyWithOption(msg string, replyOptions replyOptions) error {
	if c.replyCount >= 1 {
		c.logger.Error("Trying to reply more than once?", zap.Int("replyCount", c.replyCount))
	}
	c.replyCount += 1
	flags := discordgo.MessageFlags(0)
	config, err := c.getConfig()
	if err != nil {
		c.logger.Error("Failed to load config. Not critical - skipping.", zap.Error(err))
	}
	if replyOptions.IsPrivate || (config != nil && config.UseEphemeral) {
		flags |= discordgo.MessageFlagsEphemeral
	}
	return c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   flags,
		},
	})
}

func (c *NativeSlashHandlingContext) reply(msg string) error {
	return c.replyWithOption(msg, replyOptions{
		IsPrivate: false,
	})
}

func (c *NativeSlashHandlingContext) onError(err error) {
	c.logger.WithOptions(zap.AddCallerSkip(1)).Error(err.Error())
	if err := c.reply(utils.GenericErrorMessage(err)); err != nil {
		c.logger.Error("Failed to send generic error message.", zap.Error(err))
	}
}

func (c *NativeSlashHandlingContext) getConfig() (*models.GuildConfig, error) {
	if c.cachedConfig != nil {
		return c.cachedConfig, nil
	}
	guildID := c.i.GuildID
	config, err := c.guildConfigRepository.Get(guildID)
	if err != nil {
		return nil, err
	}
	c.cachedConfig = config
	return config, nil
}

// NativeSlashHandler are functions that are responsible for handling response and replies fully
type NativeSlashHandler = func(c *NativeSlashHandlingContext)

var (
	nativeSlashHandlerMap = map[string]NativeSlashHandler{
		"developer":               handleDeveloper,
		"generate_new_api_client": handleDeveloperCreateNewApiClient,
		"config":                  handleConfig,
	}
)

type NativeSlashHandlingParams struct {
	Session           *discordgo.Session
	InteractionCreate *discordgo.InteractionCreate
	CommandName       string
	DB                *gorm.DB
	Logger            *zap.Logger
}

func Handle(p NativeSlashHandlingParams) bool {
	handler, ok := nativeSlashHandlerMap[p.CommandName]
	if !ok {
		return false
	}
	ctx := &NativeSlashHandlingContext{
		s:                     p.Session,
		i:                     p.InteractionCreate,
		apiClientRepository:   &repositories.ApiClientRepositoryImpl{DB: p.DB},
		guildConfigRepository: &repositories.GuildConfigRepositoryImpl{DB: p.DB},
		logger:                p.Logger.Named("native"),
	}
	handler(ctx)
	return true
}
