package grocery

import (
	"context"
	"time"

	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

func (s *GroceryServiceImpl) ProcessListlessGroceries(ctx context.Context, groceries []models.GroceryEntry) error {
	if len(groceries) == 0 {
		return nil
	}

	// Queue entries to channel (non-blocking)
	queuedCount := 0
	for _, entry := range groceries {
		select {
		case s.listlessGroceriesChannel <- entry:
			queuedCount++
		default:
			// Channel is full, log warning and continue
			s.logger.Warn("ListlessGroceries channel is full, skipping entry",
				zap.Uint("entryID", entry.ID),
				zap.String("itemDesc", entry.ItemDesc))
		}
	}

	s.logger.Info("Queued listless groceries for processing",
		zap.Int("queuedCount", queuedCount),
		zap.Int("totalCount", len(groceries)))

	// Start worker if not already running
	s.startListlessGroceriesWorker(ctx)

	return nil
}

func (s *GroceryServiceImpl) startListlessGroceriesWorker(ctx context.Context) {
	if !s.workerMutex.TryLock() {
		// Worker is already running, abort immediately
		return
	}
	defer s.workerMutex.Unlock()

	// Check if worker is already running by trying to process from channel
	select {
	case entry := <-s.listlessGroceriesChannel:
		// Got an entry, worker is needed
		go s.processListlessGroceriesWorker(ctx, entry)
	default:
		// No entries in channel, no worker needed
		return
	}
}

func (s *GroceryServiceImpl) processListlessGroceriesWorker(ctx context.Context, firstEntry models.GroceryEntry) {
	s.logger.Info("Starting listless groceries worker")

	// Process the first entry
	s.processListlessEntry(ctx, firstEntry)

	// Process remaining entries from channel
	for {
		select {
		case entry := <-s.listlessGroceriesChannel:
			s.processListlessEntry(ctx, entry)
		case <-time.After(100 * time.Millisecond):
			// Channel empty for 100ms, exit worker
			s.logger.Info("Listless groceries worker exiting - channel empty")
			return
		}
	}
}

func (s *GroceryServiceImpl) processListlessEntry(ctx context.Context, entry models.GroceryEntry) {
	// Skip entries with nil GroceryListID (this is valid)
	if entry.GroceryListID == nil {
		s.logger.Debug("Skipping entry with nil GroceryListID",
			zap.Uint("entryID", entry.ID),
			zap.String("itemDesc", entry.ItemDesc))
		return
	}

	// Check if grocery list exists
	groceryList, err := s.groceryListRepo.GetByQuery(&models.GroceryList{
		ID: *entry.GroceryListID,
	})
	if err != nil {
		s.logger.Error("Error checking grocery list existence",
			zap.Uint("entryID", entry.ID),
			zap.Uint("groceryListID", *entry.GroceryListID),
			zap.Error(err))
		return
	}

	if groceryList == nil {
		// Grocery list doesn't exist, delete the entry
		if err := s.groceryEntryRepo.Delete(ctx, &entry); err != nil {
			s.logger.Error("Failed to delete orphaned grocery entry",
				zap.Uint("entryID", entry.ID),
				zap.String("itemDesc", entry.ItemDesc),
				zap.Uint("groceryListID", *entry.GroceryListID),
				zap.Error(err))
		} else {
			s.logger.Debug("Deleted orphaned grocery entry",
				zap.Uint("entryID", entry.ID),
				zap.String("itemDesc", entry.ItemDesc),
				zap.Uint("groceryListID", *entry.GroceryListID))
		}
	} else {
		// Grocery list exists, this shouldn't happen - log error
		s.logger.Error("Grocery list exists but entry was marked as listless",
			zap.Uint("entryID", entry.ID),
			zap.String("itemDesc", entry.ItemDesc),
			zap.Uint("groceryListID", *entry.GroceryListID),
			zap.String("groceryListLabel", groceryList.ListLabel))
	}
}
