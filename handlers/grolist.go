package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
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
	return nil
}

func (m *MessageHandlerContext) displayList() error {
	msgPrefix := "Here's your grocery list:"
	groceries := make([]models.GroceryEntry, 0)
	if r := m.db.Where(&models.GroceryEntry{GuildID: m.msg.GuildID}).Find(&groceries); r.Error != nil {
		return m.onError(r.Error)
	}
	if len(groceries) == 0 {
		return m.sendMessage("You have no groceries - add one through `!gro` (e.g. `!gro Toilet paper`)!")
	}
	var groceryLists []models.GroceryList
	if r := m.db.Where(&models.GroceryList{GuildID: m.msg.GuildID}).Find(&groceryLists); r.Error != nil {
		// ignore error - not considered fatal
		m.LogError(r.Error)
	}
	// group by their grocerylist
	noListGroceries := make([]models.GroceryEntry, 0)
	groupedGroceries := make(map[uint][]models.GroceryEntry)
	groceryListMap := make(map[uint]models.GroceryList)
	for _, l := range groceryLists {
		groupedGroceries[l.ID] = make([]models.GroceryEntry, 0)
		groceryListMap[l.ID] = l
	}
	for _, g := range groceries {
		if g.GroceryListID == nil {
			noListGroceries = append(noListGroceries, g)
		} else {
			l := groupedGroceries[*g.GroceryListID]
			if l == nil {
				m.LogError(
					errors.New("Unknown grocery list ID in grocery entry."),
					zap.Uint("GroceryListID", *g.GroceryListID),
				)
			} else {
				l = append(l, g)
				groupedGroceries[*g.GroceryListID] = l
			}
		}
	}
	noListGroceriesTxt := m.getGroceryListText(noListGroceries)
	labeledGroceriesTxt := ""
	for k, v := range groupedGroceries {
		matchedGroceryList := groceryListMap[k]
		label := matchedGroceryList.GetName()
		labeledGroceriesTxt += fmt.Sprintf(":shopping_cart: **%s**\n%s\n", label, m.getGroceryListText(v))
	}
	return m.sendMessage(strings.Join([]string{msgPrefix, noListGroceriesTxt, labeledGroceriesTxt}, "\n"))
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
	var fancyName *string
	if len(splitArgs) >= 3 && splitArgs[2] != "" {
		fancyName = &splitArgs[2]
	}
	var existingCount int64
	countR := m.db.Model(&models.GroceryList{}).Where(&models.GroceryList{GuildID: m.msg.GuildID, ListLabel: label}).Count(&existingCount)
	if countR.Error != nil {
		m.LogError(countR.Error)
		return m.sendMessage(msgCannotSaveNewGroceryList)
	}
	if existingCount > 0 {
		return m.sendMessage(fmt.Sprintf("Sorry, a grocery list with the label *%s* already exists for your server. Please select another label :)", label))
	}
	newGroceryList := models.GroceryList{
		GuildID:   m.msg.GuildID,
		ListLabel: label,
		FancyName: fancyName,
	}
	if r := m.db.Create(&newGroceryList); r.Error != nil {
		m.LogError(countR.Error)
		return m.sendMessage(msgCannotSaveNewGroceryList)
	}
	return m.sendMessage(fmt.Sprintf("Yay! Your new grocery list *%s* has been successfully created. Use it in a command like so to add entries to your grocery list: `gro:%s Chicken`", newGroceryList.GetName(), newGroceryList.ListLabel))
}

func (m *MessageHandlerContext) getGroceryListText(groceries []models.GroceryEntry) string {
	if len(groceries) == 0 {
		return "This grocery list is empty\n"
	}
	msg := ""
	for i, grocery := range groceries {
		msg += fmt.Sprintf("%d: %s\n", i+1, grocery.ItemDesc)
	}
	return msg
}
