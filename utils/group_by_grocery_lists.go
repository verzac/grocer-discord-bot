package utils

import "github.com/verzac/grocer-discord-bot/models"

func GroupByGroceryLists(
	groceryLists []models.GroceryList,
	groceries []models.GroceryEntry,
) (
	noListGroceries []models.GroceryEntry,
	groupedGroceries map[uint][]models.GroceryEntry,
	listlessGroceries []models.GroceryEntry,
) {
	noListGroceries = make([]models.GroceryEntry, 0)
	groupedGroceries = make(map[uint][]models.GroceryEntry)
	listlessGroceries = make([]models.GroceryEntry, 0)
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
				listlessGroceries = append(listlessGroceries, g)
			} else {
				l = append(l, g)
				groupedGroceries[*g.GroceryListID] = l
			}
		}
	}
	return
}
