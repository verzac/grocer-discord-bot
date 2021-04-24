package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
)

func (m *MessageHandler) OnBulk(argStr string) error {
	items := strings.Split(
		strings.Trim(argStr, "\n \t"),
		"\n",
	)
	toInsert := make([]models.GroceryEntry, 0, len(items))
	for _, item := range items {
		aID := m.msg.Author.ID
		cleanedItem := strings.Trim(item, " \n\t")
		if cleanedItem != "" {
			toInsert = append(toInsert, models.GroceryEntry{
				ItemDesc:    cleanedItem,
				GuildID:     m.msg.GuildID,
				UpdatedByID: &aID,
			})
		}
	}
	insertedItemsCount := int64(len(toInsert))
	if len(toInsert) > 0 {
		if err := m.checkLimit(m.msg.GuildID, int64(len(toInsert))); err != nil {
			return m.onError(err)
		}
		r := m.db.Create(&toInsert)
		if r.Error != nil {
			log.Println(m.fmtErrMsg(r.Error))
			return m.sendMessage("Hmm... Cannot save your grocery list. Please try again later :)")
		}
		insertedItemsCount = r.RowsAffected
	}
	return m.sendMessage(fmt.Sprintf("Added %d items into your list!", insertedItemsCount))
}
