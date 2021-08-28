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
	groceries, err := m.groceryEntryRepo.FindByQuery(
		&models.GroceryEntry{
			GuildID:       m.msg.GuildID,
			GroceryListID: groceryListID,
		},
	)
	if err != nil {
		return m.onError(err)
	}
	if itemIndex > len(groceries) || itemIndex < 1 {
		m.sendMessage(fmt.Sprintf("Hmm... Can't seem to find item #%d on the list :/", itemIndex))
		return m.OnList()
	}
	g := groceries[itemIndex]
	return m.sendMessage(fmt.Sprintf("Here's what you have for item #%d: %s (%s)", itemIndex, g.ItemDesc, g.GetUpdatedByString()))
}
