package handlers

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandler) OnClear() error {
	r := m.db.Delete(models.GroceryEntry{}, "guild_id = ?", m.msg.GuildID)
	if r.Error != nil {
		return m.onError(r.Error)
	}
	msg := fmt.Sprintf("Deleted %d items off your grocery list!", r.RowsAffected)
	return m.sendMessage(msg)
}
