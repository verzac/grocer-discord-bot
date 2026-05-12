package native

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/config"
	"github.com/verzac/grocer-discord-bot/handlers/slash/defaults"
	"github.com/verzac/grocer-discord-bot/services/ingredients"
	"github.com/verzac/grocer-discord-bot/utils"
	"go.uber.org/zap"
)

const ingredientsMessageMaxRunes = 1800

func handleIngredients(c *NativeSlashHandlingContext) {
	if c.i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if !config.IsN8NEnabled() {
		c.logger.Error("ingredients: disabled because n8n is disabled")
		if err := c.reply("`/ingredients` isn't ready yet. Please check back in later, or ask the team in the support server when it'll be ready. Thanks!"); err != nil {
			c.onError(err)
		}
		return
	}
	url := ""
	for _, o := range c.i.ApplicationCommandData().Options {
		if o.Name == "url" {
			url = strings.TrimSpace(o.StringValue())
			break
		}
	}
	if url == "" {
		if err := c.reply("Please provide a **url** to your recipe or video."); err != nil {
			c.onError(err)
		}
		return
	}
	listLabel := defaults.ListLabelFromSlashOptions(c.i.ApplicationCommandData().Options)

	if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	}); err != nil {
		c.logger.Error("ingredients: deferred respond failed", zap.Error(err))
		return
	}

	// Discord requires an initial interaction response within ~3s; the deferred ack above satisfies that.
	// The n8n request and FollowupMessageCreate may take ~10s+. discordgo already runs AddHandler callbacks
	// on their own goroutine when Session.SyncEvents is false (the default; see Session.handle in event.go),
	// so this blocking work does not stall the WebSocket read loop—only this handler's goroutine.

	// separate context as it is _technically_ async (the outer context may still have a timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	items, err := ingredients.Service.FetchIngredients(ctx, url)
	if err != nil {
		c.logger.Error("ingredients: fetch error", zap.Error(err))
		if _, ferr := c.s.FollowupMessageCreate(c.i.Interaction, true, &discordgo.WebhookParams{
			Content: "Could not fetch ingredients, please try again later. If the problem persists, please contact my hooman. Thanks!",
		}); ferr != nil {
			c.logger.Error("ingredients: followup after fetch error", zap.Error(ferr))
		}
		return
	}
	if len(items) == 0 {
		if _, ferr := c.s.FollowupMessageCreate(c.i.Interaction, true, &discordgo.WebhookParams{
			Content: "I didn't get any ingredients back from that link. Try a different URL or add items manually with `/grobulk`.",
		}); ferr != nil {
			c.logger.Error("ingredients: followup empty list", zap.Error(ferr))
		}
		return
	}
	cacheKey := ingredients.Service.StorePending(items, c.i.GuildID, c.i.Member.User.ID, listLabel)
	body := formatIngredientsFollowupBody(items)
	_, ferr := c.s.FollowupMessageCreate(c.i.Interaction, true, &discordgo.WebhookParams{
		Content: body,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "YES",
						Style:    discordgo.SuccessButton,
						CustomID: "ingredients_confirm:" + cacheKey,
					},
					discordgo.Button{
						Label:    "NO",
						Style:    discordgo.SecondaryButton,
						CustomID: "ingredients_cancel:" + cacheKey,
					},
				},
			},
		},
	})
	if ferr != nil {
		c.logger.Error("ingredients: followup with buttons failed", zap.Error(ferr))
	}
}

func formatIngredientsFollowupBody(items []string) string {
	itemsSection := ""
	isTruncated := false
	outFmt := `
Here are the ingredients I found from your recipe:

%s

Does this look right to you?
`
	truncatedFromIdx := 0
	outFmt = strings.TrimSpace(outFmt)

	// calculate the ingredient list section
	for i, item := range items {
		line := fmt.Sprintf("%d. %s\n", i+1, item)
		if len(itemsSection)+len(outFmt)+len(line) > ingredientsMessageMaxRunes {
			isTruncated = true
			truncatedFromIdx = i
			break
		}
		itemsSection += line
	}

	if isTruncated {
		itemsSection += fmt.Sprintf("\n**and %d other grocery items**\n", len(items)-1-truncatedFromIdx)
	}
	out := fmt.Sprintf(outFmt, itemsSection)

	return out
}

func handleIngredientsConfirm(c *NativeSlashHandlingContext) {
	if c.i.Type != discordgo.InteractionMessageComponent {
		return
	}
	key := strings.TrimSpace(c.customIDSuffix)
	if key == "" {
		if err := respondIngredientsComponentError(c.s, c.i, "Missing confirmation id."); err != nil {
			c.logger.Error("ingredients_confirm: respond", zap.Error(err))
		}
		return
	}
	if !ingredientsPendingKeyMatchesGuild(key, c.i.GuildID) {
		if err := respondIngredientsComponentError(c.s, c.i, "Something went wrong... Please try again."); err != nil {
			c.logger.Error("ingredients_confirm: guild mismatch", zap.Error(err))
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	n, err := ingredients.Service.ConfirmAndAdd(ctx, key, c.i.Member.User.ID)
	if errors.Is(err, ingredients.ErrPendingNotFound) {
		if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Sorry, this confirmation has expired. You can copy the ingredients from the list above and add them manually with `/grobulk`.",
				Components: []discordgo.MessageComponent{},
			},
		}); err != nil {
			c.logger.Error("ingredients_confirm: update expired", zap.Error(err))
		}
		return
	}
	if errors.Is(err, ingredients.ErrWrongAuthor) {
		if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Oops, that button can only be pressed by whoever ran `/ingredients`. Please try again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			c.logger.Error("ingredients_confirm: wrong author", zap.Error(err))
		}
		return
	}
	if err != nil {
		if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: utils.GenericErrorMessage(err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			c.logger.Error("ingredients_confirm: error respond", zap.Error(err))
		}
		return
	}
	msg := fmt.Sprintf("Added %d items to your grocery list!", n)
	if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    msg,
			Components: []discordgo.MessageComponent{},
		},
	}); err != nil {
		c.logger.Error("ingredients_confirm: update success but respond failed", zap.Error(err))
	}
}

func handleIngredientsCancel(c *NativeSlashHandlingContext) {
	if c.i.Type != discordgo.InteractionMessageComponent {
		return
	}
	key := strings.TrimSpace(c.customIDSuffix)
	if key == "" {
		if err := respondIngredientsComponentError(c.s, c.i, "Missing confirmation id."); err != nil {
			c.logger.Error("ingredients_cancel: respond", zap.Error(err))
		}
		return
	}
	if !ingredientsPendingKeyMatchesGuild(key, c.i.GuildID) {
		if err := respondIngredientsComponentError(c.s, c.i, "This confirmation doesn't belong to this server."); err != nil {
			c.logger.Error("ingredients_cancel: guild mismatch", zap.Error(err))
		}
		return
	}
	authorID, ok := ingredients.Service.PendingAuthorID(key)
	if ok && authorID != c.i.Member.User.ID {
		if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Those buttons are for whoever ran `/ingredients`.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			c.logger.Error("ingredients_cancel: wrong author", zap.Error(err))
		}
		return
	}
	ingredients.Service.Cancel(key)
	if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "No worries -- ingredients were not added.",
			Components: []discordgo.MessageComponent{},
		},
	}); err != nil {
		c.logger.Error("ingredients_cancel: update", zap.Error(err))
	}
}

func ingredientsPendingKeyMatchesGuild(cacheKey, guildID string) bool {
	g, _, ok := strings.Cut(cacheKey, ":")
	return ok && g == guildID
}

func respondIngredientsComponentError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
