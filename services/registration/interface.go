package registration

import (
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Service RegistrationService
)

type RegistrationService interface {
	GetRegistrationContext(guildID string) (*dto.RegistrationContext, error)
}

type RegistrationServiceImpl struct {
	guildRegistrationRepo repositories.GuildRegistrationRepository
	logger                *zap.Logger
}

func Init(db *gorm.DB, logger *zap.Logger) {
	if Service == nil {
		Service = &RegistrationServiceImpl{
			guildRegistrationRepo: &repositories.GuildRegistrationRepositoryImpl{DB: db},
			logger:                logger.Named("registration"),
		}
	}
}
