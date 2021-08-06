package handlers

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

func (m *MessageHandlerContext) OnAdd() error {
	guildID := m.msg.GuildID
	argStr := m.commandContext.ArgStr
	if argStr == "" {
		return m.sendMessage("Sorry, I need to know what you want to add to your grocery list :sweat_smile: (e.g. `!gro Chicken wings`)")
	}
	if err := m.checkLimit(guildID, 1); err != nil {
		return m.onError(err)
	}
	var groceryList *models.GroceryList
	groceryListLabel := m.commandContext.GrocerySublist
	if groceryListLabel != "" {
		if r := m.db.Where(&models.GroceryList{ListLabel: groceryListLabel, GuildID: guildID}).Take(&groceryList); r.Error != nil {
			if r.Error == gorm.ErrRecordNotFound {
				return m.sendMessage(fmt.Sprintf("Whoops, can't seem to find the grocery list labeled as *%s*", groceryListLabel))
			}
			return m.onError(r.Error)
		}
	}
	if r := m.db.Create(&models.GroceryEntry{ItemDesc: argStr, GuildID: m.msg.GuildID, UpdatedByID: &m.msg.Author.ID, GroceryList: groceryList}); r.Error != nil {
		return m.onError(r.Error)
	}
	groceryListName := "your grocery list"
	if groceryListLabel != "" && groceryList != nil {
		if groceryList.FancyName != nil {
			groceryListName = *groceryList.FancyName
		} else {
			groceryListName = groceryListLabel
		}
		groceryListName = "*" + groceryListName + "*"
	}
	err := m.sendMessage(fmt.Sprintf("Added *%s* into %s!", argStr, groceryListName))
	if err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}
