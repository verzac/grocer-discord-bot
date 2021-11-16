package handlers

import (
	"fmt"
)

func (m *MessageHandlerContext) OnClear() error {
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onGetGroceryListError(err)
	}
	rowsAffected, rErr := m.groceryEntryRepo.ClearGroceryList(groceryList, m.msg.GuildID)
	if rErr != nil {
		return m.onError(rErr)
	}
	msg := fmt.Sprintf("Deleted %d items off your grocery list!", rowsAffected)
	if err := m.sendMessage(msg); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohereWithGroceryList()
}
