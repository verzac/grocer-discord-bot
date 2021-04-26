package handlers

import (
	"errors"
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (m *MessageHandler) OnAttach() error {
	if err := m.sendMessage("Gotcha! Attaching a self-updating grocery list to the current channel. Please stand by..."); err != nil {
		return m.onError(err)
	}
	attachMsg, err := m.sess.ChannelMessageSend(m.msg.ChannelID, "Placeholder")
	if err != nil {
		return m.onError(err)
	}
	gConfig := models.GuildConfig{
		GuildID:          m.msg.GuildID,
		GrohereChannelID: &attachMsg.ChannelID,
		GrohereMessageID: &attachMsg.ID,
	}
	if r := m.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&gConfig); r.Error != nil {
		return m.onError(r.Error)
	}
	return m.onEditUpdateGrohere()
}

func (m *MessageHandler) getGrohereText() (string, error) {
	groceries := make([]models.GroceryEntry, 0)
	if r := m.db.Where(&models.GroceryEntry{GuildID: m.msg.GuildID}).Find(&groceries); r.Error != nil {
		return "", r.Error
	}
	groceryListText := m.getGroceryListText(groceries)
	var lastG *models.GroceryEntry
	for _, g := range groceries {
		if lastG == nil || lastG.UpdatedAt.Before(g.UpdatedAt) {
			lastG = &g
		}
	}
	beginningText := ":shopping_cart: **AUTO GROCERY LIST** :shopping_cart::\n"
	if lastG == nil {
		return beginningText + "You have no groceries here - hooray!", nil
	}
	lastUpdatedByText := fmt.Sprintf("Last %s\n", lastG.GetUpdatedByString())
	groHereText := fmt.Sprintf(beginningText+"%s\n%s\n", groceryListText, lastUpdatedByText)
	return groHereText, nil
}

func (m *MessageHandler) onEditUpdateGrohere() error {
	gConfig := models.GuildConfig{}
	if r := m.db.Where(&models.GuildConfig{GuildID: m.msg.GuildID}).Take(&gConfig); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			// ignore
			return nil
		} else {
			return m.onError(r.Error)
		}
	}
	if gConfig.GrohereChannelID == nil || gConfig.GrohereMessageID == nil {
		return m.onError(errors.New("Hmm... That's weird, something didn't go right when updating your !grohere grocery list. Please run !grohere again to setup a new, self-updating grocery list!"))
	}
	grohereText, err := m.getGrohereText()
	if err != nil {
		return m.onError(err)
	}
	_, err = m.sess.ChannelMessageEdit(*gConfig.GrohereChannelID, *gConfig.GrohereMessageID, grohereText)
	if err != nil {
		return m.onError(err)
	}
	return nil
}
