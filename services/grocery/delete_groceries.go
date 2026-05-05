package grocery

import (
	"context"
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

// GroceryEntriesNotFoundError is returned when one or more grocery entry IDs do not exist in the guild.
type GroceryEntriesNotFoundError struct {
	IDs []int64
}

func (e *GroceryEntriesNotFoundError) Error() string {
	return fmt.Sprintf("Grocery entries not found for IDs: %v.", e.IDs)
}

func uintSliceFromPositiveInt64IDs(ids []int64) ([]uint, error) {
	maxU := ^uint(0)
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id < 1 {
			return nil, fmt.Errorf("grocery id must be positive")
		}
		if uint64(id) > uint64(maxU) {
			return nil, fmt.Errorf("grocery id out of range: %d", id)
		}
		out = append(out, uint(id))
	}
	return out, nil
}

// DeleteGroceriesByIDs removes all entries for the given IDs in guildID. IDs may contain duplicates.
// If any ID is missing in the guild, it returns *GroceryEntriesNotFoundError and deletes nothing.
func (s *GroceryServiceImpl) DeleteGroceriesByIDs(ctx context.Context, guildID string, ids []int64) error {
	uintIDs, err := uintSliceFromPositiveInt64IDs(ids)
	if err != nil {
		return err
	}

	seen := make(map[uint]struct{}, len(uintIDs))
	uniqueIDs := make([]uint, 0, len(uintIDs))
	for _, id := range uintIDs {
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

	foundSet := make(map[uint]struct{}, len(entries))
	for i := range entries {
		foundSet[entries[i].ID] = struct{}{}
	}
	var notFound []int64
	for _, id := range uniqueIDs {
		if _, ok := foundSet[id]; !ok {
			notFound = append(notFound, int64(id))
		}
	}
	if len(notFound) > 0 {
		return &GroceryEntriesNotFoundError{IDs: notFound}
	}

	deletedIDs := make([]uint, len(entries))
	for i := range entries {
		deletedIDs[i] = entries[i].ID
	}
	if _, err := repo.DeleteByGuildAndIDs(ctx, guildID, deletedIDs); err != nil {
		return err
	}

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
