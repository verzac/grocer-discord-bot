package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
)

type getItemsToRemoveFuncType = func(args []string, groceries []models.GroceryEntry, groceryList *models.GroceryList) ([]models.GroceryEntry, error)

func (m *MessageHandlerContext) OnRemove() error {
	argStr := m.commandContext.ArgStr
	args := strings.Split(argStr, " ")
	if len(args) == 0 {
		return nil
	}
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onError(err)
	}
	var groceryListID *uint
	if groceryList != nil {
		groceryListID = &groceryList.ID
	}
	groceries, err := m.groceryEntryRepo.FindByQueryWithConfig(
		&models.GroceryEntry{
			GuildID:       m.msg.GuildID,
			GroceryListID: groceryListID,
		},
		repositories.GroceryEntryQueryOpts{
			IsStrongNilForGroceryListID: true,
		},
	)
	if err != nil {
		return m.onError(err)
	}
	if len(groceries) == 0 {
		msg := fmt.Sprintf("Whoops, you do not have any items in %s.", groceryList.GetName())
		return m.sendMessage(msg)
	}
	var getItemsToRemoveFunc getItemsToRemoveFuncType
	if _, err := strconv.Atoi(args[0]); err == nil {
		getItemsToRemoveFunc = getItemsToRemoveWithIndex
	} else {
		getItemsToRemoveFunc = getItemsToRemoveWithName
	}
	toDelete, err := getItemsToRemoveFunc(args, groceries, groceryList)
	if err != nil {
		return m.sendMessage(err.Error())
	}
	if rDel := m.db.Delete(toDelete); rDel.Error != nil {
		return m.onError(rDel.Error)
	}
	if err := m.sendMessage(fmt.Sprintf("Deleted %s off %s!", prettyItems(toDelete), groceryList.GetName())); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}

func getItemsToRemoveWithName(args []string, groceries []models.GroceryEntry, groceryList *models.GroceryList) ([]models.GroceryEntry, error) {
	searchStr := strings.Join(args, " ")
	toDelete := make([]models.GroceryEntry, 1)
	for _, g := range groceries {
		// locales and non-standard unicodes can really mess things up
		if strings.Contains(g.ItemDesc, searchStr) || strings.Contains(strings.ToLower(g.ItemDesc), strings.ToLower(searchStr)) {
			toDelete[0] = g
			return toDelete, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Whoops, I cannot find %s in %s.", searchStr, groceryList.GetName()))
}

func getItemsToRemoveWithIndex(args []string, groceries []models.GroceryEntry, groceryList *models.GroceryList) ([]models.GroceryEntry, error) {
	itemIndexes, err := getItemIndexes(args)
	if err != nil {
		return nil, err
	}
	toDelete := make([]models.GroceryEntry, 0, len(itemIndexes))
	missingIndexes := make([]int, 0, len(itemIndexes))
	for _, itemIndex := range itemIndexes {
		// validate
		if itemIndex > len(groceries) {
			missingIndexes = append(missingIndexes, itemIndex)
		} else if len(missingIndexes) == 0 {
			toDelete = append(toDelete, groceries[itemIndex-1])
		}
	}
	if len(missingIndexes) > 0 {
		return nil, errors.New(fmt.Sprintf(
			"Whoops, we can't seem to find the following item(s) in %s: %s",
			groceryList.GetName(),
			prettyItemIndexList(missingIndexes),
		))
	}
	return toDelete, nil
}

func getItemIndexes(args []string) ([]int, error) {
	itemIndexes := make([]int, 0)
	for _, itemIndexStr := range args {
		if itemIndexStr != "" {
			itemIndex, err := toItemIndex(itemIndexStr)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf(
					"Hmm... I can't seem to understand that number (%s) - they don't seem valid to me. Mind trying again?",
					itemIndexStr,
				))
			}
			itemIndexes = append(itemIndexes, itemIndex)
		}
	}
	return itemIndexes, nil
}
