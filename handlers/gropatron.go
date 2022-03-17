package handlers

import (
	"fmt"
	"time"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandlerContext) OnPatron() error {
	argStr := m.commandContext.ArgStr
	switch argStr {
	case "register":
		return m.onRegister()
	case "deregister":
		return m.onDeregister()
	default:
		return m.reply(`
Hmm... Not sure what you were looking for. Here are my available commands:
` + "`!gropatron register`" + ` - registers your account's Patreon benefit for this server.
` + "`!gropatron deregister`" + ` - deregisters your account's Patreon benefit for this server, and allows you to register other servers (if you've hit your limit).
		`)
	}
}

func (m *MessageHandlerContext) getEntitlement() (*models.RegistrationEntitlement, error) {
	authorID := m.commandContext.AuthorID
	username := m.commandContext.AuthorUsername
	discriminator := m.commandContext.AuthorUsernameDiscriminator
	return m.registrationEntitlementRepo.GetActive(&models.RegistrationEntitlement{UserID: &authorID, Username: &username, UsernameDiscriminator: &discriminator})
}

func (m *MessageHandlerContext) onRegister() error {
	guildID := m.commandContext.GuildID
	entitlement, err := m.getEntitlement()
	if err != nil {
		return m.onError(err)
	}
	if entitlement == nil {
		return m.reply("Oops, you do not have the entitlement to register this server.")
	}
	guildRegistrationsByUser, err := m.guildRegistrationRepo.FindByQuery(&models.GuildRegistration{RegistrationEntitlementID: entitlement.ID})
	if len(guildRegistrationsByUser) >= entitlement.MaxRedemption {
		return m.reply(fmt.Sprintf("Oops, you already have registered %d server under your account (max: %d). You can deregister servers from your account using `/gropatron deregister`.", len(guildRegistrationsByUser), entitlement.MaxRedemption))
	}
	for _, r := range guildRegistrationsByUser {
		if r.GuildID == guildID {
			return m.reply("Whoops, you've already registered this server under your account - no further actions needed! :tada:")
		}
	}
	if err := m.guildRegistrationRepo.Save(&models.GuildRegistration{
		GuildID:                   guildID,
		RegistrationEntitlementID: entitlement.ID,
		ExpiresAt:                 entitlement.ExpiresAt,
	}); err != nil {
		return m.onError(err)
	}
	return m.reply(":tada: Yay! You've successfully registered your server to your account. You can now enjoy GroceryBot's extra benefits on this server.")
}

func (m *MessageHandlerContext) onDeregister() error {
	guildID := m.commandContext.GuildID
	entitlement, err := m.getEntitlement()
	if err != nil {
		return m.onError(err)
	}
	guildRegistrationsByUser, err := m.guildRegistrationRepo.FindByQuery(&models.GuildRegistration{GuildID: guildID, RegistrationEntitlementID: entitlement.ID})
	if err != nil {
		return m.onError(err)
	}
	if len(guildRegistrationsByUser) == 0 {
		return m.reply("You haven't registered your benefits for this server - no actions needed!")
	}
	now := time.Now()
	for i := range guildRegistrationsByUser {
		guildRegistrationsByUser[i].ExpiresAt = &now
	}
	if err := m.guildRegistrationRepo.Put(guildRegistrationsByUser...); err != nil {
		return m.onError(err)
	}
	return m.reply(":wave: You've successfully removed your account's benefits from this server. Please feel free to run `/gropatron register` again to register your account!")
}
