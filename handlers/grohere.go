package handlers

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	errCannotEditMsgWhenRemovingGrolist  = errors.New("Failed to edit message when grocery list was removed.")
	errCannotEditMsgWhenReplacingGrohere = errors.New("Failed to edit /grohere message upon replacement.")
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
	if err := m.reply(fmt.Sprintf("Gotcha! Attaching a self-updating grocery list for **%s** to the current channel. Please stand by...", groceryList.GetName())); err != nil {
		return m.onError(err)
	}
	guildID := m.commandContext.GuildID
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
		_, err := m.sess.ChannelMessageEdit(record.GrohereChannelID, record.GrohereMessageID, fmt.Sprintf(":shopping_cart: %s\n*This /grohere message has been replaced in another message.*", groceryList.GetName()))
		if err != nil {
			m.LogError(errCannotEditMsgWhenReplacingGrohere, zap.Any("DiscordErr", err))
		}
	}
	if len(grohereRecords) > 0 {
		if r := m.db.Delete(grohereRecords); r.Error != nil {
			return m.onError(r.Error)
		}
	}
	attachMsg, err := m.sess.ChannelMessageSend(m.commandContext.ChannelID, "Placeholder")
	if err != nil || attachMsg == nil {
		m.GetLogger().Error("Unable to attach a message to the channel for /grohere.", zap.Error(err))
		return m.sendDirectMessage("Oops, I can't seem to attach the grocery list through `grohere`. Have you added the \"Send Message\" permission for me in your server / channel?", m.commandContext.AuthorID)
	}
	grohereRecord := models.GrohereRecord{
		GuildID:          m.commandContext.GuildID,
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
	if err := m.reply("Gotcha! Attaching a self-updating grocery list to the current channel. Please stand by..."); err != nil {
		return m.onError(err)
	}
	attachMsg, err := m.sess.ChannelMessageSend(m.commandContext.ChannelID, "Placeholder")
	if err != nil {
		return m.onError(err)
	}
	gConfig := models.GuildConfig{
		GuildID:          m.commandContext.GuildID,
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

func (m *MessageHandlerContext) onEditUpdateGrohere() error {
	if err := m.groceryService.UpdateGuildGrohere(context.Background(), m.commandContext.GuildID); err != nil {
		return m.onError(err)
	}
	return nil
}

// onEditUpdateGrohereWithGroceryList is for commands that support grocery-specific context
func (m *MessageHandlerContext) onEditUpdateGrohereWithGroceryList() error {
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onError(err)
	}
	if err := m.groceryService.OnGroceryListEdit(context.Background(), groceryList, m.commandContext.GuildID); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}

func (m *MessageHandlerContext) onListRemoveGrohereRecord(groceryList *models.GroceryList) error {
	record := models.GrohereRecord{}
	if r := m.db.Model(&models.GrohereRecord{}).Where("guild_id = ? AND grocery_list_id = ?", m.commandContext.GuildID, groceryList.ID).Take(&record); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			// nothing needs to be cleaned up
			return nil
		}
		return m.onError(r.Error)
	}
	_, err := m.sess.ChannelMessageEdit(record.GrohereChannelID, record.GrohereMessageID, fmt.Sprintf(":shopping_cart: %s\n *:wave: This grocery list has been deleted. Type `/grohere:<your-new-grocery-list>` to get a self-updating message for your grocery list!*", groceryList.GetName()))
	if err != nil {
		m.LogError(errCannotEditMsgWhenRemovingGrolist, zap.String("DiscordErr", err.Error()))
	}
	if r := m.db.Delete(record); r.Error != nil {
		return m.onError(r.Error)
	}
	return nil
}
