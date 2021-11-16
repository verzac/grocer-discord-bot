package handlers

import (
	"fmt"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandlerContext) OnEdit() error {
	argStr := m.commandContext.ArgStr
	argTokens := strings.SplitN(argStr, " ", 2)
	if len(argTokens) != 2 {
		return m.sendMessage(fmt.Sprintf("Oops, I can't seem to understand you. Perhaps try typing **!groedit 1 Whatever you want the name of this entry to be**?"))
	}
	itemIndex, err := toItemIndex(argTokens[0])
	if err != nil {
		return m.sendMessage(err.Error())
	}
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onGetGroceryListError(err)
	}
	newItemDesc := argTokens[1]
	guildID := m.msg.GuildID
	var groceryListID *uint
	if groceryList != nil {
		groceryListID = &groceryList.ID
	}
	g, err := m.groceryEntryRepo.GetByItemIndex(
		&models.GroceryEntry{
			GuildID:       guildID,
			GroceryListID: groceryListID,
		},
		itemIndex,
	)
	if err != nil {
		return m.onError(err)
	}
	if g == nil {
		return m.onItemNotFound(itemIndex)
	}
	g.ItemDesc = newItemDesc
	g.UpdatedByID = &m.msg.Author.ID
	if err := m.groceryEntryRepo.Put(g); err != nil {
		m.LogError(err)
		return m.sendMessage("Welp, something went wrong while saving. Please try again :)")
	}
	if err := m.sendMessage(fmt.Sprintf("Updated item #%d on your grocery list to *%s*", itemIndex, g.ItemDesc)); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohereWithGroceryList()
}
