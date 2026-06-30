package grocery

import (
	"context"

	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

func (s *GroceryServiceImpl) ValidateGroceryListLimit(ctx context.Context, registrationContext *dto.RegistrationContext, guildID string) (limitOk bool, limit int, err error) {
	limit = registrationContext.MaxGroceryListsPerServer
	existingNamedListCount, err := s.groceryListRepo.WithContext(ctx).Count(&models.GroceryList{GuildID: guildID})
	if err != nil {
		return false, limit, err
	}

	// Add one for the new named list and one for the implicit default list, which is not stored in grocery_lists.
	totalListCount := int(existingNamedListCount) + 2
	if totalListCount > limit {
		s.logger.Warn("max grocery list limit exceeded.",
			zap.String("guildID", guildID),
			zap.Any("registrationContext", registrationContext),
			zap.Int("Limit", limit),
			zap.Int("TotalListCount", totalListCount),
		)
		return false, limit, nil
	}
	return true, limit, nil
}
