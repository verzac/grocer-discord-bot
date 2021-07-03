package handlers

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandlerContext) OnRemove(argStr string) error {
	args := strings.Split(argStr, " ")
	groceryList := make([]models.GroceryEntry, 0)
	rFind := m.db.Where("guild_id = ?", m.msg.GuildID).Find(&groceryList)
	if rFind.Error != nil {
		return m.onError(rFind.Error)
	}
	if rFind.RowsAffected == 0 {
		msg := fmt.Sprintf("Whoops, you do not have any items in your grocery list.")
		return m.sendMessage(msg)
	}
	toDelete, err := getItemsToRemove(args, groceryList)
	if err != nil {
		return m.sendMessage(err.Error())
	}
	if rDel := m.db.Delete(toDelete); rDel.Error != nil {
		return m.onError(rDel.Error)
	}
	if err := m.sendMessage(fmt.Sprintf("Deleted %s off your grocery list!", prettyItems(toDelete))); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}

func getItemsToRemove(args []string, groceryList []models.GroceryEntry) ([]models.GroceryEntry, error) {
	itemIndexes, err := getItemIndexes(args)
	if err != nil {
		return nil, err
	}
	toDelete := make([]models.GroceryEntry, 0, len(itemIndexes))
	missingIndexes := make([]int, 0, len(itemIndexes))
	for _, itemIndex := range itemIndexes {
		// validate
		if itemIndex > len(groceryList) {
			missingIndexes = append(missingIndexes, itemIndex)
		} else if len(missingIndexes) == 0 {
			toDelete = append(toDelete, groceryList[itemIndex-1])
		}
	}
	if len(missingIndexes) > 0 {
		return nil, errors.New(fmt.Sprintf(
			"Whoops, we can't seem to find the following item(s): %s",
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
