package native

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

const (
	waitlistIosModalCustomID      = "waitlist_ios_submit"
	waitlistIosEmailInputCustomID = "waitlist_ios_email"
	waitlistIosNameInputCustomID  = "waitlist_ios_name"
	waitlistIosInteractionTimeout = 10 * time.Second
	waitlistIosModalTitleMaxRunes = 45
	waitlistIosModalIntroText     = "Thanks for your interest - we'll let you know once we've released GroceryBot App on iOS!\n\n[Android Play Store Link](https://play.google.com/store/apps/details?id=net.grocerybot.app)\n\n_We will not use your email collected here for any other purposes._"
)

var handleWaitlistIos NativeSlashHandler = func(c *NativeSlashHandlingContext) {
	if c.i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if c.i.Member == nil {
		if err := c.reply("This command can only be used in a server."); err != nil {
			c.onError(err)
		}
		return
	}
	options := c.i.ApplicationCommandData().Options
	if len(options) != 1 || options[0].Type != discordgo.ApplicationCommandOptionSubCommand || options[0].Name != "ios" {
		c.onError(errMissingSubcommand)
		return
	}

	modalTitle := "iOS app waitlist"
	if len([]rune(modalTitle)) > waitlistIosModalTitleMaxRunes {
		modalTitle = string([]rune(modalTitle)[:waitlistIosModalTitleMaxRunes])
	}

	data := &discordgo.InteractionResponseData{
		CustomID: waitlistIosModalCustomID,
		Title:    modalTitle,
		Components: []discordgo.MessageComponent{
			discordgo.TextDisplay{Content: waitlistIosModalIntroText},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    waitlistIosEmailInputCustomID,
						Label:       "Email",
						Style:       discordgo.TextInputShort,
						Placeholder: "you@example.com",
						Required:    true,
						MinLength:   3,
						MaxLength:   256,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    waitlistIosNameInputCustomID,
						Label:       "Name (optional)",
						Style:       discordgo.TextInputShort,
						Placeholder: "Name we'll refer to you by in the email (optional)",
						Required:    false,
						MaxLength:   100,
					},
				},
			},
		},
	}
	if err := c.s.InteractionRespond(c.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: data,
	}); err != nil {
		c.logger.Error("waitlist_ios: open modal failed", zap.Error(err))
	}
}

var handleWaitlistIosSubmit NativeSlashHandler = func(c *NativeSlashHandlingContext) {
	if c.i.Type != discordgo.InteractionModalSubmit {
		return
	}
	if c.i.Member == nil {
		return
	}
	data := c.i.ModalSubmitData()
	if data.CustomID != waitlistIosModalCustomID {
		return
	}

	email, name, ok := parseWaitlistIosModalValues(data.Components)
	if !ok {
		if err := respondWaitlistIosEphemeral(c.s, c.i, "Please enter a valid email address."); err != nil {
			c.logger.Error("waitlist_ios_submit: invalid email respond failed", zap.Error(err))
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), waitlistIosInteractionTimeout)
	defer cancel()

	userID := c.i.Member.User.ID
	existing, err := c.waitlistIosRepository.FindByDiscordUserID(ctx, userID)
	if err != nil {
		c.logger.Error("waitlist_ios_submit: find failed", zap.Error(err))
		if err := respondWaitlistIosEphemeral(c.s, c.i, "Something went wrong saving your signup. Please try again later."); err != nil {
			c.logger.Error("waitlist_ios_submit: error respond failed", zap.Error(err))
		}
		return
	}
	var msg string
	if existing != nil {
		previousEmail := existing.Email
		if err := c.waitlistIosRepository.UpdateByDiscordUserID(ctx, userID, email, name); err != nil {
			c.logger.Error("waitlist_ios_submit: update failed", zap.Error(err))
			if err := respondWaitlistIosEphemeral(c.s, c.i, "Something went wrong saving your signup. Please try again later."); err != nil {
				c.logger.Error("waitlist_ios_submit: update error respond failed", zap.Error(err))
			}
			return
		}
		msg = fmt.Sprintf(
			"You're still on the list! We've replaced the email we had for your Discord account (%s) with the one you just submitted.",
			previousEmail,
		)
	} else {
		entry := &models.WaitlistIos{
			DiscordUserID: userID,
			Email:         email,
			Name:          name,
		}
		if err := c.waitlistIosRepository.Create(ctx, entry); err != nil {
			c.logger.Error("waitlist_ios_submit: create failed", zap.Error(err))
			if err := respondWaitlistIosEphemeral(c.s, c.i, "Something went wrong saving your signup. Please try again later."); err != nil {
				c.logger.Error("waitlist_ios_submit: create error respond failed", zap.Error(err))
			}
			return
		}
		msg = "You're on the list! We'll let you know once GroceryBot for iOS is available on the App Store."
	}
	if err := respondWaitlistIosEphemeral(c.s, c.i, msg); err != nil {
		c.logger.Error("waitlist_ios_submit: success respond failed", zap.Error(err))
	}
}

func parseWaitlistIosModalValues(components []discordgo.MessageComponent) (email string, name string, ok bool) {
	var emailRaw, nameRaw string
	for _, row := range components {
		ar, okRow := row.(*discordgo.ActionsRow)
		if !okRow || len(ar.Components) == 0 {
			continue
		}
		ti, okTi := ar.Components[0].(*discordgo.TextInput)
		if !okTi {
			continue
		}
		switch ti.CustomID {
		case waitlistIosEmailInputCustomID:
			emailRaw = ti.Value
		case waitlistIosNameInputCustomID:
			nameRaw = ti.Value
		}
	}
	email = strings.TrimSpace(emailRaw)
	name = strings.TrimSpace(nameRaw)
	if email == "" || !strings.Contains(email, "@") {
		return "", "", false
	}
	return email, name, true
}

func respondWaitlistIosEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
