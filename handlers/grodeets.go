package handlers

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

func (m *MessageHandler) OnDetail(argStr string) error {
	itemIndex, err := toItemIndex(argStr)
	if err != nil {
		return m.sendMessage(err.Error())
	}
	g := models.GroceryEntry{}
	r := m.db.Where("guild_id = ?", m.msg.GuildID).Offset(itemIndex - 1).First(&g)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			m.sendMessage(fmt.Sprintf("Hmm... Can't seem to find item #%d on the list :/", itemIndex))
			return m.OnList()
		}
		return m.onError(r.Error)
	}
	if r.RowsAffected == 0 {
		msg := fmt.Sprintf("Cannot find item with index %d!", itemIndex)
		return m.sendMessage(msg)
	}
	return m.sendMessage(fmt.Sprintf("Here's what you have for item #%d: %s (%s)", itemIndex, g.ItemDesc, g.GetUpdatedByString()))
}
