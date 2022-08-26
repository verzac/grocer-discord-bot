package registration

import (
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/repositories"
	"gorm.io/gorm"
)

type RegistrationService interface {
	GetRegistrationContext(guildID string) (*dto.RegistrationContext, error)
}

type RegistrationServiceImpl struct {
	guildRegistrationRepo repositories.GuildRegistrationRepository
}

func NewRegistrationService(db *gorm.DB) RegistrationService {
	return &RegistrationServiceImpl{
		guildRegistrationRepo: &repositories.GuildRegistrationRepositoryImpl{DB: db},
	}
}
