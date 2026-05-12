package ingredients

import (
	"context"
	"errors"

	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Service IngredientsService

	ErrPendingNotFound = errors.New("pending ingredients not found or expired")
	ErrWrongAuthor     = errors.New("this confirmation belongs to another user")
)

type IngredientsService interface {
	FetchIngredients(ctx context.Context, url string) ([]string, error)
	StorePending(ingredients []string, guildID, authorID, listLabel string) string
	PendingAuthorID(cacheKey string) (authorID string, ok bool)
	ConfirmAndAdd(ctx context.Context, cacheKey, authorID string) (addedCount int, err error)
	Cancel(cacheKey string)
}

type IngredientsServiceImpl struct {
	db               *gorm.DB
	logger           *zap.Logger
	memCache         *pendingCache
	groceryListRepo  repositories.GroceryListRepository
	groceryEntryRepo repositories.GroceryEntryRepository
}

func Init(db *gorm.DB, logger *zap.Logger) {
	if Service == nil {
		Service = &IngredientsServiceImpl{
			db:       db,
			logger:   logger.Named("ingredients"),
			memCache: newPendingCache(),
			groceryListRepo: &repositories.GroceryListRepositoryImpl{
				DB: db,
			},
			groceryEntryRepo: &repositories.GroceryEntryRepositoryImpl{
				DB: db,
			},
		}
	}
}
