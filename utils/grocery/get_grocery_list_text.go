package groceryutils

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
)

func GetGroceryListText(groceries []models.GroceryEntry, groceryList *models.GroceryList) string {
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
