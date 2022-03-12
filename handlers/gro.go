package handlers

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
)

func (m *MessageHandlerContext) OnAdd() error {
	guildID := m.commandContext.GuildID
	argStr := m.commandContext.ArgStr
	if argStr == "" {
		return m.reply("Sorry, I need to know what you want to add to your grocery list :sweat_smile: (e.g. `!gro Chicken wings`)")
	}
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onGetGroceryListError(err)
	}
	limitOk, groceryEntryLimit, err := m.ValidateGroceryEntryLimit(guildID, 1)
	if err != nil {
		return m.onError(err)
	}
	if !limitOk {
		return m.reply(msgOverLimit(groceryEntryLimit))
	}
	rErr := m.groceryEntryRepo.AddToGroceryList(
		groceryList,
		[]models.GroceryEntry{
			{
				ItemDesc:    argStr,
				GuildID:     m.commandContext.GuildID,
				UpdatedByID: &m.commandContext.AuthorID,
				GroceryList: groceryList,
			},
		},
		guildID)
	if rErr != nil {
		switch rErr.ErrCode {
		case repositories.ErrCodeValidationError:
			return m.reply(rErr.Error())
		default:
			return m.onError(rErr)
		}
	}
	groceryListName := "your grocery list"
	if groceryList != nil {
		groceryListName = groceryList.GetName()
	}
	err = m.reply(fmt.Sprintf("Added *%s* into %s!", argStr, groceryListName))
	if err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohereWithGroceryList()
}
