package handlers

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandlerContext) OnDetail() error {
	argStr := m.commandContext.ArgStr
	itemIndex, err := toItemIndex(argStr)
	if err != nil {
		return m.sendMessage(err.Error())
	}
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onGetGroceryListError(err)
	}
	var groceryListID *uint
	if groceryList != nil {
		groceryListID = &groceryList.ID
	}
	g, err := m.groceryEntryRepo.GetByItemIndex(
		&models.GroceryEntry{
			GuildID:       m.msg.GuildID,
			GroceryListID: groceryListID,
		},
		itemIndex,
	)
	if err != nil {
		return m.onError(err)
	}
	return m.sendMessage(fmt.Sprintf("Here's what you have for item #%d: %s (%s)", itemIndex, g.ItemDesc, g.GetUpdatedByString()))
}
