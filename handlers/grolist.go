package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"github.com/verzac/grocer-discord-bot/utils"
	"go.uber.org/zap"
)

const (
	msgCannotSaveNewGroceryList = "Whoops, can't seem to save your new grocery list. Please try again later!"
)

func (m *MessageHandlerContext) OnList() error {
	if m.commandContext.ArgStr == "" {
		return m.displayList()
	}
	if strings.HasPrefix(m.commandContext.ArgStr, "new ") {
		return m.newList()
	}
	if strings.HasPrefix(m.commandContext.ArgStr, "delete ") {
		return m.deleteList()
	}
	return nil
}

func (m *MessageHandlerContext) displayList() error {
	msgPrefix := "Here's your grocery list:"
	groceries, err := m.groceryEntryRepo.FindByQuery(&models.GroceryEntry{GuildID: m.msg.GuildID})
	if err != nil {
		return m.onError(err)
	}
	groceryLists, err := m.groceryListRepo.FindByQuery(&models.GroceryList{GuildID: m.msg.GuildID})
	if err != nil {
		return m.onError(err)
	}
	textBody, err := m.getDisplayListText(groceryLists, groceries)
	if err != nil {
		return m.onError(err)
	}
	if len(groceries) == 0 && len(groceryLists) == 0 {
		msgPrefix = ""
	}
	return m.sendMessage(strings.Join([]string{msgPrefix, textBody}, "\n"))
}

func (m *MessageHandlerContext) getDisplayListText(groceryLists []models.GroceryList, groceries []models.GroceryEntry) (string, error) {
	// group by their grocerylist
	if len(groceryLists) == 0 && len(groceries) == 0 {
		return getNoGroceryText(""), nil
	}
	noListGroceries, groupedGroceries, listlessGroceries := utils.GroupByGroceryLists(groceryLists, groceries)
	for _, g := range listlessGroceries {
		m.LogError(
			errors.New("Unknown grocery list ID in grocery entry."),
			zap.Uint("GroceryListID", *g.GroceryListID),
		)
	}
	noListGroceriesTxt := getGroceryListText(noListGroceries, nil)
	labeledGroceriesTxt := ""
	for _, groceryList := range groceryLists {
		g := groupedGroceries[groceryList.ID]
		label := groceryList.ListLabel
		fancyName := groceryList.FancyName
		groceryListText := ""
		if fancyName != nil {
			groceryListText = fmt.Sprintf("**%s** (%s)", *fancyName, label)
		} else {
			groceryListText = fmt.Sprintf("**%s**", label)
		}
		labeledGroceriesTxt += fmt.Sprintf(":shopping_cart: %s\n%s\n", groceryListText, getGroceryListText(g, &groceryList))
	}
	return strings.Join([]string{noListGroceriesTxt, labeledGroceriesTxt}, "\n"), nil
}

func (m *MessageHandlerContext) newList() error {
	splitArgs := strings.SplitN(m.commandContext.ArgStr, " ", 3)
	if len(splitArgs) < 2 {
		return m.sendMessage("Sorry, I need to know what you'd like to label your new grocery list as. For example, you can type `!grolist new amazon` to make a grocery list with the label `amazon`.")
	}
	// if len(splitArgs) > 2 {
	// 	return m.sendMessage("Sorry, grocery list labels cannot contain any spaces. You can fix this by using \"-\" or \".\" to replace your spaces (e.g. `!grolist new morning market` -> `!grolist new morning-market`). PS: In the future, I'll be able to give your grocery lists custom names.")
	// }
	label := splitArgs[1]
	var fancyName string
	if len(splitArgs) >= 3 && splitArgs[2] != "" {
		fancyName = splitArgs[2]
	}
	newGroceryList, err := m.groceryListRepo.CreateGroceryList(m.msg.GuildID, label, fancyName)
	if err != nil {
		switch err {
		case repositories.ErrGroceryListDuplicate:
			return m.sendMessage(fmt.Sprintf("Sorry, a grocery list with the label *%s* already exists for your server. Please select another label :)", label))
		default:
			m.LogError(err)
			return m.sendMessage(msgCannotSaveNewGroceryList)
		}
	}
	if err := m.sendMessage(fmt.Sprintf("Yay! Your new grocery list *%s* has been successfully created. Use it in a command like so to add entries to your grocery list: `gro:%s Chicken`", newGroceryList.GetName(), newGroceryList.ListLabel)); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}

func fmtErrGroceryListNotFound(label string) string {
	return fmt.Sprintf("Whoops, I cannot seem to find a grocery list with the name %s... Could you please try again?", label)
}

func (m *MessageHandlerContext) deleteList() error {
	splitArgs := strings.SplitN(m.commandContext.ArgStr, " ", 3)
	if len(splitArgs) < 2 {
		return m.sendMessage("Sorry, I need to know which grocery list you'd like to delete. For example, you can type `!grolist delete amazon` to delete a grocery list with the label `amazon`.")
	}
	label := splitArgs[1]
	groceryList, err := m.groceryListRepo.GetByQuery(&models.GroceryList{ListLabel: label})
	if err != nil {
		return m.onError(err)
	}
	if groceryList == nil {
		return m.sendMessage(fmtErrGroceryListNotFound(label))
	}
	count, rErr := m.groceryEntryRepo.GetCount(&models.GroceryEntry{GuildID: m.msg.GuildID, GroceryListID: &groceryList.ID})
	if rErr != nil {
		return m.onError(rErr)
	}
	if count > 0 {
		return m.sendMessage(fmt.Sprintf("Oops, you still have %d groceries in *%s*", count, groceryList.ListLabel))
	}
	if err := m.groceryListRepo.Delete(groceryList); err != nil {
		switch err {
		case repositories.ErrGroceryListNotFound:
			return m.sendMessage(fmtErrGroceryListNotFound(label))
		default:
			return m.onError(err)
		}
	}
	return m.sendMessage(fmt.Sprintf("Successfully deleted **%s**! Feel free to make new ones with `!grolist new`!", label))
}

func getGroceryListText(groceries []models.GroceryEntry, groceryList *models.GroceryList) string {
	if groceryList != nil && len(groceries) == 0 {
		label := ""
		if groceryList != nil {
			label = ":" + groceryList.ListLabel
		}
		return getNoGroceryText(label)
	}
	msg := ""
	for i, grocery := range groceries {
		msg += fmt.Sprintf("%d: %s\n", i+1, grocery.ItemDesc)
	}
	return msg
}

func getNoGroceryText(label string) string {
	return fmt.Sprintf("You have no groceries - add one through `!gro` (e.g. `!gro%s Toilet paper`)!\n", label)
}
