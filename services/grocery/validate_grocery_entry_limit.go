package grocery

import (
	"context"

	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

func (s *GroceryServiceImpl) ValidateGroceryEntryLimit(ctx context.Context, registrationContext *dto.RegistrationContext, guildID string, newItemCount int) (limitOk bool, limit int, err error) {
	limit = registrationContext.MaxGroceryEntriesPerServer
	count, err := s.groceryEntryRepo.WithContext(ctx).GetCount(&models.GroceryEntry{GuildID: guildID})
	if err != nil {
		return false, limit, err
	}
	return s.ValidateGroceryEntryLimitUsingTotalCount(ctx, registrationContext, guildID, int(count)+newItemCount)
}

func (s *GroceryServiceImpl) ValidateGroceryEntryLimitUsingTotalCount(ctx context.Context, registrationContext *dto.RegistrationContext, guildID string, totalItemCount int) (limitOk bool, limit int, err error) {
	limit = registrationContext.MaxGroceryEntriesPerServer
	if totalItemCount > limit {
		s.logger.Warn("max grocery list limit exceeded.",
			zap.String("guildID", guildID),
			zap.Any("registrationContext", registrationContext),
			zap.Int("Limit", limit),
			zap.Int("NewItemCount", totalItemCount),
		)
		return false, limit, nil
	}
	return true, limit, nil
}
