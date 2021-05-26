package handlers

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandlerContext) OnClear() error {
	r := m.db.Delete(models.GroceryEntry{}, "guild_id = ?", m.msg.GuildID)
	if r.Error != nil {
		return m.onError(r.Error)
	}
	msg := fmt.Sprintf("Deleted %d items off your grocery list!", r.RowsAffected)
	if err := m.sendMessage(msg); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}
