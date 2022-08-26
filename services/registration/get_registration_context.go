package registration

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/config"
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/models"
)

func (s *RegistrationServiceImpl) GetRegistrationContext(guildID string) (*dto.RegistrationContext, error) {
	registrationContext := dto.RegistrationContext{
		MaxGroceryListsPerServer:   config.GetDefaultMaxGroceryListsPerServer(),
		MaxGroceryEntriesPerServer: config.GetDefaultMaxGroceryEntriesPerServer(),
		IsDefault:                  true,
	}
	registrationsOwnersMention := []string{}
	registrations, err := s.guildRegistrationRepo.FindByQuery(&models.GuildRegistration{GuildID: guildID})
	if err != nil {
		// return default
		return &registrationContext, err
	}
	for _, r := range registrations {
		if currentMaxLists := r.RegistrationEntitlement.RegistrationTier.MaxGroceryList; currentMaxLists != nil && *currentMaxLists > registrationContext.MaxGroceryListsPerServer {
			registrationContext.MaxGroceryListsPerServer = *currentMaxLists
		}
		if currentMaxEntries := r.RegistrationEntitlement.RegistrationTier.MaxGroceryEntry; currentMaxEntries != nil && *currentMaxEntries > registrationContext.MaxGroceryEntriesPerServer {
			registrationContext.MaxGroceryEntriesPerServer = *currentMaxEntries
		}
		if registrationUserID := r.RegistrationEntitlement.UserID; registrationUserID != nil {
			registrationsOwnersMention = append(registrationsOwnersMention, fmt.Sprintf("<@%s>", *r.RegistrationEntitlement.UserID))
		}
		registrationContext.IsDefault = false
	}
	registrationContext.RegistrationsOwnersMention = registrationsOwnersMention
	return &registrationContext, nil
}
