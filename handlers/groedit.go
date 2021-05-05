package handlers

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

func (m *MessageHandler) OnEdit(argStr string) error {
	argTokens := strings.SplitN(argStr, " ", 2)
	if len(argTokens) != 2 {
		return m.sendMessage(fmt.Sprintf("Oops, I can't seem to understand you. Perhaps try typing **!groedit 1 Whatever you want the name of this entry to be**?"))
	}
	itemIndex, err := toItemIndex(argTokens[0])
	if err != nil {
		return m.sendMessage(err.Error())
	}
	newItemDesc := argTokens[1]
	g := models.GroceryEntry{}
	fr := m.db.Where("guild_id = ?", m.msg.GuildID).Offset(itemIndex - 1).First(&g)
	if fr.Error != nil {
		if errors.Is(fr.Error, gorm.ErrRecordNotFound) {
			m.sendMessage(fmtItemNotFoundErrorMsg(itemIndex))
			return m.OnList()
		}
		return m.onError(fr.Error)
	}
	if fr.RowsAffected == 0 {
		msg := fmt.Sprintf("Cannot find item with index %d!", itemIndex)
		return m.sendMessage(msg)
	}
	g.ItemDesc = newItemDesc
	g.UpdatedByID = &m.msg.Author.ID
	if sr := m.db.Save(g); sr.Error != nil {
		log.Println(m.FmtErrMsg(sr.Error))
		return m.sendMessage("Welp, something went wrong while saving. Please try again :)")
	}
	if err := m.sendMessage(fmt.Sprintf("Updated item #%d on your grocery list to *%s*", itemIndex, g.ItemDesc)); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}
