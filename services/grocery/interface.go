package grocery

import (
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Service GroceryService
)

type GroceryService interface {
	ValidateGroceryEntryLimit(registrationContext *dto.RegistrationContext, guildID string, newItemCount int) (limitOk bool, limit int, err error)
}

type GroceryServiceImpl struct {
	groceryEntryRepo repositories.GroceryEntryRepository
	// grohereRepo      repositories.GrohereRecordRepository
	logger *zap.Logger
}

func Init(db *gorm.DB, logger *zap.Logger) {
	if Service == nil {
		Service = &GroceryServiceImpl{
			// grohereRepo:      &repositories.GrohereRecordRepositoryImpl{DB: db},
			groceryEntryRepo: &repositories.GroceryEntryRepositoryImpl{DB: db},
			logger:           logger.Named("grocery"),
		}
	}
}
