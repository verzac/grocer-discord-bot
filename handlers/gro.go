package handlers

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
)

func (m *MessageHandlerContext) OnAdd() error {
	guildID := m.msg.GuildID
	argStr := m.commandContext.ArgStr
	if argStr == "" {
		return m.sendMessage("Sorry, I need to know what you want to add to your grocery list :sweat_smile: (e.g. `!gro Chicken wings`)")
	}
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		if err == errGroceryListNotFound {
			return m.sendMessage(m.commandContext.FmtErrInvalidGroceryList())
		} else {
			return m.onError(err)
		}
	}
	rErr := m.groceryEntryRepo.AddToGroceryList(
		groceryList,
		[]models.GroceryEntry{
			{
				ItemDesc:    argStr,
				GuildID:     m.msg.GuildID,
				UpdatedByID: &m.msg.Author.ID,
				GroceryList: groceryList,
			},
		},
		guildID)
	if rErr != nil {
		switch rErr.ErrCode {
		case repositories.ErrCodeValidationError:
			return m.sendMessage(rErr.Error())
		default:
			return m.onError(rErr)
		}
	}
	groceryListName := "your grocery list"
	if groceryList != nil {
		groceryListName = groceryList.GetName()
	}
	err = m.sendMessage(fmt.Sprintf("Added *%s* into %s!", argStr, groceryListName))
	if err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}
