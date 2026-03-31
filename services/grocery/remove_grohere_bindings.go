package grocery

import (
	"context"
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

// RemoveGrohereBindingsForList edits list-scoped !grohere messages to a deleted-list notice, then removes grohere_records rows.
func (s *GroceryServiceImpl) RemoveGrohereBindingsForList(ctx context.Context, groceryList *models.GroceryList, guildID string) error {
	if groceryList == nil {
		return nil
	}
	records, err := s.grohereRepo.FindByQuery(&models.GrohereRecord{
		GuildID:       guildID,
		GroceryListID: groceryList.GetID(),
	})
	if err != nil {
		return err
	}
	deletedMsg := fmt.Sprintf(
		":shopping_cart: %s\n *:wave: This grocery list has been deleted. Type `!grohere:<your-new-grocery-list>` to get a self-updating message for your grocery list!*",
		groceryList.GetName(),
	)
	for i := range records {
		rec := &records[i]
		if _, editErr := s.sess.ChannelMessageEdit(rec.GrohereChannelID, rec.GrohereMessageID, deletedMsg); editErr != nil {
			s.logger.Error(
				"Failed to edit message when grocery list was removed.",
				zap.Error(editErr),
			)
		}
		if delErr := s.grohereRepo.Delete(rec); delErr != nil {
			return delErr
		}
	}
	return nil
}
