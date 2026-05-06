package grocery

import (
	"context"
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

// GroceryEntriesNotFoundError is returned when one or more grocery entry IDs do not exist in the guild.
type GroceryEntriesNotFoundError struct {
	IDs []uint
}

func (e *GroceryEntriesNotFoundError) Error() string {
	return fmt.Sprintf("Grocery entries not found for IDs: %v.", e.IDs)
}

// DeleteGroceriesByIDs removes all entries for the given IDs in guildID. IDs may contain duplicates.
// If any ID is missing in the guild, it returns *GroceryEntriesNotFoundError and deletes nothing.
func (s *GroceryServiceImpl) DeleteGroceriesByIDs(ctx context.Context, guildID string, ids []uint) error {
	// process and dedupe IDs
	seen := make(map[uint]struct{}, len(ids))
	uniqueIDs := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	repo := s.groceryEntryRepo.WithContext(ctx)
	entries, err := repo.FindByGuildAndIDs(ctx, guildID, uniqueIDs)
	if err != nil {
		return err
	}

	if len(entries) != len(uniqueIDs) {
		// some IDs were not found, causing a mismatch
		// figure out which IDs were not found
		foundSet := make(map[uint]struct{}, len(entries))
		for i := range entries {
			foundSet[entries[i].ID] = struct{}{}
		}
		notFound := make([]uint, 0)
		for _, id := range uniqueIDs {
			if _, ok := foundSet[id]; !ok {
				notFound = append(notFound, id)
			}
		}
		return &GroceryEntriesNotFoundError{IDs: notFound}
	}

	if _, err := repo.DeleteByGuildAndIDs(ctx, guildID, uniqueIDs); err != nil {
		return err
	}

	// process grohere
	listIDSet := make(map[uint]struct{})
	hasListless := false
	for i := range entries {
		if entries[i].GroceryListID == nil || *entries[i].GroceryListID == 0 {
			hasListless = true
			continue
		}
		listIDSet[*entries[i].GroceryListID] = struct{}{}
	}

	if len(listIDSet) > 0 {
		allLists, lErr := s.groceryListRepo.FindByQuery(&models.GroceryList{GuildID: guildID})
		if lErr != nil {
			return lErr
		}
		listsByID := make(map[uint]*models.GroceryList, len(allLists))
		for i := range allLists {
			listsByID[allLists[i].ID] = &allLists[i]
		}
		for listID := range listIDSet {
			gl := listsByID[listID]
			if err := s.OnGroceryListEdit(ctx, gl, guildID); err != nil {
				s.logger.Error("Failed to run OnGroceryListEdit", zap.Error(err))
			}
		}
	}
	if hasListless {
		if err := s.OnGroceryListEdit(ctx, nil, guildID); err != nil {
			s.logger.Error("Failed to run OnGroceryListEdit", zap.Error(err))
		}
	}

	return nil
}
