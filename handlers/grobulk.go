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
	toInsert := make([]models.GroceryEntry, len(items))
	for i, item := range items {
		aID := m.msg.Author.ID
		toInsert[i] = models.GroceryEntry{
			ItemDesc:    strings.Trim(item, " \n\t"),
			GuildID:     m.msg.GuildID,
			UpdatedByID: &aID,
		}
	}
	if err := m.checkLimit(m.msg.GuildID, int64(len(toInsert))); err != nil {
		return m.onError(err)
	}
	r := m.db.Create(&toInsert)
	if r.Error != nil {
		log.Println(m.fmtErrMsg(r.Error))
		return m.sendMessage("Hmm... Cannot save your grocery list. Please try again later :)")
	}
	return m.sendMessage(fmt.Sprintf("Added %d items into your list!", r.RowsAffected))
}
