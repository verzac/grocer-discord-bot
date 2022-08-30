package grocery

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	groceryutils "github.com/verzac/grocer-discord-bot/utils/grocery"
	"go.uber.org/zap"
)

var (
	ErrCannotUpdateGrohere = errors.New("cannot edit attached message/channel: deleting !grohere entry")
)

func (s *GroceryServiceImpl) OnGroceryListEdit(ctx context.Context, groceryList *models.GroceryList, guildID string) error {
	groceryListID := groceryList.GetID()
	grohereRecords, err := s.grohereRepo.FindByQueryWithConfig(&models.GrohereRecord{GuildID: guildID, GroceryListID: groceryListID}, repositories.GrohereRecordQueryOpts{
		IsStrongNilForGroceryListID: true,
	})
	if err != nil {
		return err
	}
	if len(grohereRecords) == 0 {
		return s.UpdateGuildGrohere(ctx, guildID)
	}
	// marshal text
	groceryLists := make([]models.GroceryList, 0, 1)
	if groceryList != nil {
		groceryLists = append(groceryLists, *groceryList)
	}
	groceries, err := s.groceryEntryRepo.FindByQueryWithConfig(&models.GroceryEntry{
		GuildID:       guildID,
		GroceryListID: groceryListID,
	}, repositories.GroceryEntryQueryOpts{
		IsStrongNilForGroceryListID: true,
	})
	if err != nil {
		return err
	}
	grohereText, listlessGroceries := groceryutils.GetGrohereText(groceryLists, groceries, true)
	if len(listlessGroceries) > 0 {
		s.logger.Error("Detected groceries without matching grocery list IDs.", zap.Any("listlessGroceries", listlessGroceries))
	}
	// if (groceryList == nil && count > 0) || (groceryList != nil && count > 1) {
	// 	grohereText += fmt.Sprintf("\nand %d other grocery lists (use `!grohere all` to get a self-updating list for all groceries, or use `!grolist all` to display them).", count)
	// }
	// send the message
	record := grohereRecords[0]
	grohereChannelID := record.GrohereChannelID
	grohereMsgID := record.GrohereMessageID
	_, err = s.sess.ChannelMessageEdit(grohereChannelID, grohereMsgID, grohereText)
	if err != nil {
		if discordErr, ok := err.(*discordgo.RESTError); ok {
			s.logger.Error(
				"Cannot edit attached message/channel: deleting !grohere entry",
				zap.Any("DiscordErr", discordErr),
			)
			// clear grohere entry as it refers to an unknown channel
			if err := s.grohereRepo.Delete(&record); err != nil {
				return err
			}
		}
		return ErrCannotUpdateGrohere
	}
	return s.UpdateGuildGrohere(ctx, guildID)
}

func (s *GroceryServiceImpl) UpdateGuildGrohere(ctx context.Context, guildID string) error {
	gConfig, err := s.guildConfigRepo.Get(guildID)
	if err != nil {
		return err
	}
	if gConfig == nil || gConfig.GrohereChannelID == nil || gConfig.GrohereMessageID == nil {
		return nil
	}
	groceryLists, err := s.groceryListRepo.FindByQuery(&models.GroceryList{GuildID: guildID})
	if err != nil {
		return err
	}
	groceries, err := s.groceryEntryRepo.FindByQuery(&models.GroceryEntry{GuildID: guildID})
	if err != nil {
		return err
	}
	grohereText, listlessGroceries := groceryutils.GetGrohereText(groceryLists, groceries, false)
	if len(listlessGroceries) > 0 {
		s.logger.Error("listlessGroceries detected.", zap.Any("listlessGroceries", listlessGroceries))
	}
	_, err = s.sess.ChannelMessageEdit(*gConfig.GrohereChannelID, *gConfig.GrohereMessageID, grohereText)
	if err != nil {
		if discordErr, ok := err.(*discordgo.RESTError); ok {
			s.logger.Error("Cannot edit attached message/channel: deleting !grohere entry", zap.Any("discordErr", discordErr))
			// clear grohere entry as it refers to an unknown channel
			gConfig.GrohereChannelID = nil
			gConfig.GrohereMessageID = nil
			if err := s.guildConfigRepo.Put(gConfig); err != nil {
				return err
			}
		}
		return ErrCannotUpdateGrohere
	}
	return nil
}
