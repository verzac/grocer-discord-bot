package handlers

import (
	"fmt"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandlerContext) OnBulk() error {
	argStr := m.commandContext.ArgStr
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onGetGroceryListError(err)
	}
	items := strings.Split(
		strings.Trim(argStr, "\n \t"),
		"\n",
	)
	toInsert := make([]models.GroceryEntry, 0, len(items))
	for _, item := range items {
		aID := m.msg.Author.ID
		cleanedItem := strings.Trim(item, " \n\t")
		if cleanedItem != "" {
			toInsert = append(toInsert, models.GroceryEntry{
				ItemDesc:    cleanedItem,
				GuildID:     m.msg.GuildID,
				UpdatedByID: &aID,
			})
		}
	}
	insertedItemsCount := len(toInsert)
	if len(toInsert) > 0 {
		rErr := m.groceryEntryRepo.AddToGroceryList(groceryList, toInsert, m.msg.GuildID)
		if rErr != nil {
			return m.onError(rErr)
		}
	}
	listLabel := "your list"
	if groceryList != nil {
		listLabel = groceryList.GetName()
	}
	if err := m.sendMessage(fmt.Sprintf("Added %d items into %s!", insertedItemsCount, listLabel)); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}
