package groceryutils

import (
	"fmt"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/utils"
)

func GetDisplayListText(groceryLists []models.GroceryList, groceries []models.GroceryEntry) (string, []models.GroceryEntry) {
	// group by their grocerylist
	if len(groceryLists) == 0 && len(groceries) == 0 {
		return getNoGroceryText(""), make([]models.GroceryEntry, 0)
	}
	noListGroceries, groupedGroceries, listlessGroceries := utils.GroupByGroceryLists(groceryLists, groceries)
	noListGroceriesTxt := GetGroceryListText(noListGroceries, nil)
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
		labeledGroceriesTxt += fmt.Sprintf(":shopping_cart: %s\n%s\n", groceryListText, GetGroceryListText(g, &groceryList))
	}
	return strings.Join([]string{noListGroceriesTxt, labeledGroceriesTxt}, "\n"), listlessGroceries
}
