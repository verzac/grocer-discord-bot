package handlers

import "github.com/verzac/grocer-discord-bot/models"

func (m *MessageHandlerContext) OnReset() error {
	// if err := m.reply("Deleting all data for this server from my database... Please stand-by... :robot:"); err != nil {
	// 	return m.onError(err)
	// }
	if r := m.db.Delete(models.GroceryEntry{}, "guild_id = ?", m.commandContext.GuildID); r.Error != nil {
		return m.onError(r.Error)
	}
	if r := m.db.Delete(models.GuildConfig{}, "guild_id = ?", m.commandContext.GuildID); r.Error != nil {
		return m.onError(r.Error)
	}
	if r := m.db.Delete(models.GroceryList{}, "guild_id = ?", m.commandContext.GuildID); r.Error != nil {
		return m.onError(r.Error)
	}
	if r := m.db.Delete(models.GrohereRecord{}, "guild_id = ?", m.commandContext.GuildID); r.Error != nil {
		return m.onError(r.Error)
	}
	if r := m.db.Delete(models.GuildRegistration{}, "guild_id = ?", m.commandContext.GuildID); r.Error != nil {
		return m.onError(r.Error)
	}
	if err := m.reply(":wave: I've successfully deleted all of your data from my database! (p.s. you may need to set up commands such as /grohere again)"); err != nil {
		return m.onError(err)
	}
	return nil
}
