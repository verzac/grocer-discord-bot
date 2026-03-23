package slash

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/monitoring"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func modalSubmitCustomID(i *discordgo.InteractionCreate) (cid string, ok bool) {
	defer func() {
		if recover() != nil {
			cid = ""
			ok = false
		}
	}()
	if i == nil || i.Interaction == nil {
		return "", false
	}
	if i.Type != discordgo.InteractionModalSubmit {
		return "", false
	}
	return i.ModalSubmitData().CustomID, true
}

func grocerySublistLabelForHash(lists []models.GroceryList, hash string) (listLabel string, ok bool) {
	if hash == groceryListLabelHash("") {
		return "", true
	}
	for i := range lists {
		if groceryListLabelHash(lists[i].ListLabel) == hash {
			return lists[i].ListLabel, true
		}
	}
	return "", false
}

func handleGroremoveRawInteractionCreate(
	sess *discordgo.Session,
	e *discordgo.Event,
	db *gorm.DB,
	logger *zap.Logger,
	grobotVersion string,
	cw *cloudwatch.CloudWatch,
) {
	defer handlers.Recover(logger)
	if e.Type != "INTERACTION_CREATE" {
		return
	}
	parsed, isGroremove, err := parseGroremoveModalSubmitFromInteractionRaw(e.RawData)
	if err != nil {
		logger.Debug("groremove raw parse skipped.", zap.Error(err))
		return
	}
	if !isGroremove {
		return
	}

	if parsed.Member == nil {
		if err := sess.InteractionRespond(&discordgo.Interaction{
			ID:    parsed.InteractionID,
			Token: parsed.Token,
			Type:  discordgo.InteractionModalSubmit,
		}, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "I do not currently support commands through my DM. Please put me in a server and I'll work magic for you!",
			},
		}); err != nil {
			logger.Error("Cannot respond to groremove DM interaction.", zap.Error(err))
		}
		return
	}

	authorID := ""
	username := ""
	discriminator := ""
	if parsed.Member != nil && parsed.Member.User != nil {
		authorID = parsed.Member.User.ID
		username = parsed.Member.User.Username
		discriminator = parsed.Member.User.Discriminator
	} else if parsed.User != nil {
		authorID = parsed.User.ID
		username = parsed.User.Username
		discriminator = parsed.User.Discriminator
	}
	if authorID == "" {
		logger.Error("groremove modal submit missing author.")
		return
	}

	listRepo := &repositories.GroceryListRepositoryImpl{DB: db}
	lists, err := listRepo.FindByQuery(&models.GroceryList{GuildID: parsed.GuildID})
	if err != nil {
		logger.Error("groremove modal could not load grocery lists.", zap.Error(err))
		return
	}
	listLabel, known := grocerySublistLabelForHash(lists, parsed.ListHash)
	if !known {
		if err := sess.InteractionRespond(&discordgo.Interaction{
			ID:    parsed.InteractionID,
			Token: parsed.Token,
			Type:  discordgo.InteractionModalSubmit,
		}, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Whoops, I could not match that grocery list anymore. Try `/groremove` again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			logger.Error("Cannot respond to groremove unknown list.", zap.Error(err))
		}
		return
	}

	values := dedupeStringsPreserveOrder(parsed.Values)
	if len(values) == 0 {
		if err := sess.InteractionRespond(&discordgo.Interaction{
			ID:    parsed.InteractionID,
			Token: parsed.Token,
			Type:  discordgo.InteractionModalSubmit,
		}, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No items selected. Choose one or more checkboxes, or use `/grobulk` to edit your list.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			logger.Error("Cannot respond to groremove empty selection.", zap.Error(err))
		}
		return
	}

	argStr := strings.Join(values, " ")
	intr := &discordgo.Interaction{
		ID:        parsed.InteractionID,
		Token:     parsed.Token,
		Type:      discordgo.InteractionModalSubmit,
		GuildID:   parsed.GuildID,
		ChannelID: parsed.ChannelID,
		Member:    parsed.Member,
		User:      parsed.User,
	}

	commandContext := &handlers.CommandContext{
		Command:                     handlers.CmdGroRemove,
		GrocerySublist:              listLabel,
		ArgStr:                      argStr,
		GuildID:                     parsed.GuildID,
		AuthorID:                    authorID,
		ChannelID:                   parsed.ChannelID,
		AuthorUsername:              username,
		AuthorUsernameDiscriminator: discriminator,
		CommandSourceType:           handlers.CommandSourceSlashCommand,
		Interaction:                 intr,
	}

	cmdLogger := logger.Named("handler").With(
		zap.String("GuildID", parsed.GuildID),
		zap.String("SlashCommandName", "groremove-modal-submit"),
	)
	cm := monitoring.NewCommandMetric(cw, handlers.CmdGroRemove, cmdLogger)
	defer cm.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	handler := handlers.NewHandler(ctx, sess, commandContext, db, grobotVersion, cmdLogger)
	if err := handler.Handle(); err != nil {
		cmdLogger.Error("Unable to handle groremove modal submit.", zap.Any("Error", err))
		if err := sess.InteractionRespond(intr, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Whoops, something went wrong! My hoomans will fix that ASAP (usually within 24h) - sorry for the inconvenience!",
			},
		}); err != nil {
			handler.LogError(err)
		}
	}
}
