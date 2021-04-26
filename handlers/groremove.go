package handlers

import (
	"fmt"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandler) OnRemove(argStr string) error {
	itemIndexes := make([]int, 0)
	for _, itemIndexStr := range strings.Split(argStr, " ") {
		if itemIndexStr != "" {
			itemIndex, err := toItemIndex(itemIndexStr)
			if err != nil {
				return m.sendMessage(err.Error())
			}
			itemIndexes = append(itemIndexes, itemIndex)
		}
	}
	groceryList := make([]models.GroceryEntry, 0)
	rFind := m.db.Where("guild_id = ?", m.msg.GuildID).Find(&groceryList)
	if rFind.Error != nil {
		return m.onError(rFind.Error)
	}
	if rFind.RowsAffected == 0 {
		msg := fmt.Sprintf("Whoops, you do not have any items in your grocery list.")
		return m.sendMessage(msg)
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
		return m.sendMessage(fmt.Sprintf(
			"Whoops, we can't seem to find the following item(s): %s",
			prettyItemIndexList(missingIndexes),
		))
	}
	if rDel := m.db.Delete(toDelete); rDel.Error != nil {
		return m.onError(rDel.Error)
	}
	if err := m.sendMessage(fmt.Sprintf("Deleted %s off your grocery list!", prettyItems(toDelete))); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}
