package handlers

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandler) OnList() error {
	msg := "Here's your grocery list:\n"
	groceries := make([]models.GroceryEntry, 0)
	if r := m.db.Where(&models.GroceryEntry{GuildID: m.msg.GuildID}).Find(&groceries); r.Error != nil {
		return m.onError(r.Error)
	}
	groceryListText := m.getGroceryListText(groceries)
	return m.sendMessage(msg + groceryListText)
}

func (m *MessageHandler) getGroceryListText(groceries []models.GroceryEntry) string {
	msg := ""
	for i, grocery := range groceries {
		msg += fmt.Sprintf("%d: %s\n", i+1, grocery.ItemDesc)
	}
	return msg
}
