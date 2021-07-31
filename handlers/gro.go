package handlers

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandlerContext) OnAdd() error {
	argStr := m.commandContext.ArgStr
	if argStr == "" {
		return m.sendMessage("Sorry, I need to know what you want to add to your grocery list :sweat_smile: (e.g. `!gro Chicken wings`)")
	}
	if err := m.checkLimit(m.msg.GuildID, 1); err != nil {
		return m.onError(err)
	}
	if r := m.db.Create(&models.GroceryEntry{ItemDesc: argStr, GuildID: m.msg.GuildID, UpdatedByID: &m.msg.Author.ID}); r.Error != nil {
		return m.onError(r.Error)
	}
	err := m.sendMessage(fmt.Sprintf("Added *%s* into your grocery list!", argStr))
	if err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}
