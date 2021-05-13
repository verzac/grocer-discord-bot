package handlers

import "github.com/verzac/grocer-discord-bot/models"

func (m *MessageHandler) OnReset() error {
	if err := m.sendMessage("Deleting all data for this server from my database... Please stand-by... :robot:"); err != nil {
		return m.onError(err)
	}
	if r := m.db.Delete(models.GroceryEntry{}, "guild_id = ?", m.msg.GuildID); r.Error != nil {
		return m.onError(r.Error)
	}
	if r := m.db.Delete(models.GuildConfig{}, "guild_id = ?", m.msg.GuildID); r.Error != nil {
		return m.onError(r.Error)
	}
	if err := m.sendMessage(":wave: I've successfully deleted all of your data from my database! (p.s. you may need to set up commands such as !grohere again)"); err != nil {
		return m.onError(err)
	}
	return nil
}