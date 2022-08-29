package grocery

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Service GroceryService
)

type GroceryService interface {
	ValidateGroceryEntryLimit(ctx context.Context, registrationContext *dto.RegistrationContext, guildID string, newItemCount int) (limitOk bool, limit int, err error)
	OnGroceryListEdit(ctx context.Context, groceryList *models.GroceryList, guildID string) error
}

type GroceryServiceImpl struct {
	groceryEntryRepo repositories.GroceryEntryRepository
	grohereRepo      repositories.GrohereRecordRepository
	guildConfigRepo  repositories.GuildConfigRepository
	logger           *zap.Logger
	groceryListRepo  repositories.GroceryListRepository
	sess             *discordgo.Session
}

func Init(db *gorm.DB, logger *zap.Logger, sess *discordgo.Session) {
	if Service == nil {
		Service = &GroceryServiceImpl{
			grohereRepo:      &repositories.GrohereRecordRepositoryImpl{DB: db},
			groceryEntryRepo: &repositories.GroceryEntryRepositoryImpl{DB: db},
			guildConfigRepo:  &repositories.GuildConfigRepositoryImpl{DB: db},
			groceryListRepo:  &repositories.GroceryListRepositoryImpl{DB: db},
			logger:           logger.Named("grocery"),
			sess:             sess,
		}
	}
}
