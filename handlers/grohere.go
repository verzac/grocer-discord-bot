package handlers

import (
	"github.com/pkg/errors"

	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (m *MessageHandlerContext) OnAttach() error {
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

func (m *MessageHandlerContext) getGrohereText() (string, error) {
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
	lastUpdatedByText := fmt.Sprintf("Last updated by <@%s>\n", *lastG.UpdatedByID)
	groHereText := fmt.Sprintf(beginningText+"%s\n%s\n", groceryListText, lastUpdatedByText)
	return groHereText, nil
}

func (m *MessageHandlerContext) onEditUpdateGrohere() error {
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
		return nil
	}
	grohereText, err := m.getGrohereText()
	if err != nil {
		return m.onError(err)
	}
	_, err = m.sess.ChannelMessageEdit(*gConfig.GrohereChannelID, *gConfig.GrohereMessageID, grohereText)
	if err != nil {
		if discordErr, ok := err.(*discordgo.RESTError); ok {
			log.Println(m.FmtErrMsg(errors.Wrap(discordErr, "Cannot edit attached message/channel: deleting !grohere entry")))
			// clear grohere entry as it refers to an unknown channel
			gConfig.GrohereChannelID = nil
			gConfig.GrohereMessageID = nil
			m.db.Save(&gConfig)
			err := m.sendMessage("_Psst, I can't seem to edit the !grohere message I attached. If this was not intended, please use !grohere again!_")
			if err != nil {
				log.Println(m.FmtErrMsg(err))
			}
		} else {
			return m.onError(err)
		}
	}
	return nil
}
