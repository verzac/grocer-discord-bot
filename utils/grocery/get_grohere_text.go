package groceryutils

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
)

func GetGrohereText(groceryLists []models.GroceryList, groceries []models.GroceryEntry, isSingleList bool) (string, []models.GroceryEntry) {
	displayText, listlessGroceries := GetDisplayListText(groceryLists, groceries)
	var lastG *models.GroceryEntry
	for _, g := range groceries {
		if lastG == nil || lastG.UpdatedAt.Before(g.UpdatedAt) {
			lastG = &g
		}
	}
	beginningText := ":shopping_cart: **AUTO GROCERY LIST** :shopping_cart::\n"
	if isSingleList && len(groceryLists) != 0 {
		// the assumption here is that if isSingleList && groceryLists is populated, then they'd have their prefixes with the list label ready, so we don't need the original prefix
		beginningText = ""
	}
	lastUpdatedByText := ""
	if lastG != nil && lastG.UpdatedByID != nil {
		lastUpdatedByText = fmt.Sprintf("Last updated by <@%s>\n", *lastG.UpdatedByID)
	}
	groHereText := beginningText + displayText + lastUpdatedByText
	return groHereText, listlessGroceries
}
