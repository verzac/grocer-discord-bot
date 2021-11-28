package handlers

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	errCannotEditMsg                     = errors.New("Cannot edit attached message/channel: deleting !grohere entry")
	errCannotEditMsgWhenRemovingGrolist  = errors.New("Failed to edit message when grocery list was removed.")
	errCannotEditMsgWhenReplacingGrohere = errors.New("Failed to edit !grohere message upon replacement.")
)

func (m *MessageHandlerContext) OnAttach() error {
	if m.commandContext.ArgStr == "all" {
		return m.onAttachAll()
	} else if m.commandContext.ArgStr == "" {
		return m.onAttachList()
	}
	return nil
}

func (m *MessageHandlerContext) onAttachList() error {
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onGetGroceryListError(err)
	}
	if err := m.sendMessage(fmt.Sprintf("Gotcha! Attaching a self-updating grocery list for **%s** to the current channel. Please stand by...", groceryList.GetName())); err != nil {
		return m.onError(err)
	}
	guildID := m.msg.GuildID
	groceryListID := groceryList.GetID()
	q := m.db.Where(&models.GrohereRecord{GuildID: guildID, GroceryListID: groceryListID})
	if groceryListID == nil {
		q = q.Where("grocery_list_id IS NULL")
	}
	grohereRecords := make([]models.GrohereRecord, 0)
	if r := q.Find(&grohereRecords); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		return m.onError(r.Error)
	}
	for _, record := range grohereRecords {
		// there's usually only one, but you never know
		_, err := m.sess.ChannelMessageEdit(record.GrohereChannelID, record.GrohereMessageID, fmt.Sprintf(":shopping_cart: %s\n*This !grohere message has been replaced in another message.*", groceryList.GetName()))
		if err != nil {
			m.LogError(errCannotEditMsgWhenReplacingGrohere, zap.Any("DiscordErr", err))
		}
	}
	if len(grohereRecords) > 0 {
		if r := m.db.Delete(grohereRecords); r.Error != nil {
			return m.onError(r.Error)
		}
	}
	attachMsg, err := m.sess.ChannelMessageSend(m.msg.ChannelID, "Placeholder")
	grohereRecord := models.GrohereRecord{
		GuildID:          m.msg.GuildID,
		GrohereChannelID: attachMsg.ChannelID,
		GrohereMessageID: attachMsg.ID,
		GroceryListID:    groceryList.GetID(),
	}
	if r := m.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&grohereRecord); r.Error != nil {
		return m.onError(r.Error)
	}
	return m.onEditUpdateGrohereWithGroceryList()
}

func (m *MessageHandlerContext) onAttachAll() error {
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

func (m *MessageHandlerContext) getGrohereText(groceryLists []models.GroceryList, groceries []models.GroceryEntry, isSingleList bool) string {
	displayText := m.getDisplayListText(groceryLists, groceries)
	var lastG *models.GroceryEntry
	for _, g := range groceries {
		if lastG == nil || lastG.UpdatedAt.Before(g.UpdatedAt) {
			lastG = &g
		}
	}
	beginningText := ":shopping_cart: **AUTO GROCERY LIST** :shopping_cart::\n"
	if isSingleList && len(groceryLists) != 0 {
		// the assumption here is that if isSingleList && groceryLists is populated, then they'd have their prefixes with the list label ready, so we don't need the original prefix
		beginningText = ""
	}
	lastUpdatedByText := ""
	if lastG != nil {
		lastUpdatedByText = fmt.Sprintf("Last updated by <@%s>\n", *lastG.UpdatedByID)
	}
	groHereText := beginningText + displayText + lastUpdatedByText
	return groHereText
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
	guildID := m.msg.GuildID
	groceryLists, err := m.groceryListRepo.FindByQuery(&models.GroceryList{GuildID: guildID})
	if err != nil {
		return m.onError(err)
	}
	groceries, err := m.groceryEntryRepo.FindByQuery(&models.GroceryEntry{GuildID: guildID})
	if err != nil {
		return m.onError(err)
	}
	grohereText := m.getGrohereText(groceryLists, groceries, false)
	_, err = m.sess.ChannelMessageEdit(*gConfig.GrohereChannelID, *gConfig.GrohereMessageID, grohereText)
	if err != nil {
		if discordErr, ok := err.(*discordgo.RESTError); ok {
			m.LogError(errors.Wrap(discordErr, "Cannot edit attached message/channel: deleting !grohere entry"))
			// clear grohere entry as it refers to an unknown channel
			gConfig.GrohereChannelID = nil
			gConfig.GrohereMessageID = nil
			if r := m.db.Save(&gConfig); r.Error != nil {
				m.LogError(r.Error)
			}
			err := m.sendMessage("_Psst, I can't seem to edit the !grohere message I attached. If this was not intended, please use !grohere again!_")
			if err != nil {
				m.LogError(err)
			}
		} else {
			return m.onError(err)
		}
	}
	return nil
}

// onEditUpdateGrohereWithGroceryList is for commands that support grocery-specific context
func (m *MessageHandlerContext) onEditUpdateGrohereWithGroceryList() error {
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onGetGroceryListError(err)
	}
	groceryListID := groceryList.GetID()
	q := m.db.Where(&models.GrohereRecord{GuildID: m.msg.GuildID, GroceryListID: groceryListID})
	if groceryListID == nil {
		q = q.Where("grocery_list_id IS NULL")
	}
	record := models.GrohereRecord{}
	if r := q.Take(&record); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			// there is no record that needs to be updated, move onto the next part
			return m.onEditUpdateGrohere()
		}
		return m.onError(r.Error)
	}
	// marshal text
	guildID := m.msg.GuildID
	groceryLists := make([]models.GroceryList, 0, 1)
	if groceryList != nil {
		groceryLists = append(groceryLists, *groceryList)
	}
	groceries, err := m.groceryEntryRepo.FindByQueryWithConfig(&models.GroceryEntry{
		GuildID:       guildID,
		GroceryListID: groceryListID,
	}, repositories.GroceryEntryQueryOpts{
		IsStrongNilForGroceryListID: true,
	})
	if err != nil {
		return m.onError(err)
	}
	grohereText := m.getGrohereText(groceryLists, groceries, true)
	// if (groceryList == nil && count > 0) || (groceryList != nil && count > 1) {
	// 	grohereText += fmt.Sprintf("\nand %d other grocery lists (use `!grohere all` to get a self-updating list for all groceries, or use `!grolist all` to display them).", count)
	// }
	// send the message
	grohereChannelID := record.GrohereChannelID
	grohereMsgID := record.GrohereMessageID
	_, err = m.sess.ChannelMessageEdit(grohereChannelID, grohereMsgID, grohereText)
	if err != nil {
		if discordErr, ok := err.(*discordgo.RESTError); ok {
			m.LogError(
				errCannotEditMsg,
				zap.Any("DiscordErr", discordErr),
			)
			// clear grohere entry as it refers to an unknown channel
			if r := m.db.Delete(&record); r.Error != nil {
				m.LogError(r.Error)
			}
			err := m.sendMessage(fmt.Sprintf("_Psst, I can't seem to edit the !grohere message I attached for %s. If this was not intended, please use !grohere%s again!_", groceryList.GetName(), groceryList.GetLabelSuffix()))
			if err != nil {
				m.LogError(err)
			}
		} else {
			return m.onError(err)
		}
	}
	return m.onEditUpdateGrohere()
}

func (m *MessageHandlerContext) onListRemoveGrohereRecord(groceryList *models.GroceryList) error {
	record := models.GrohereRecord{}
	if r := m.db.Model(&models.GrohereRecord{}).Where("guild_id = ? AND grocery_list_id = ?", m.msg.GuildID, groceryList.ID).Take(&record); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			// nothing needs to be cleaned up
			return nil
		}
		return m.onError(r.Error)
	}
	_, err := m.sess.ChannelMessageEdit(record.GrohereChannelID, record.GrohereMessageID, fmt.Sprintf(":shopping_cart: %s\n *:wave: This grocery list has been deleted. Type `!grohere:<your-new-grocery-list>` to get a self-updating message for your grocery list!*", groceryList.GetName()))
	if err != nil {
		m.LogError(errCannotEditMsgWhenRemovingGrolist, zap.String("DiscordErr", err.Error()))
	}
	if r := m.db.Delete(record); r.Error != nil {
		return m.onError(r.Error)
	}
	return nil
}
