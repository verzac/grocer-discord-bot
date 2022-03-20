package repositories

import (
	"errors"
	"time"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ RegistrationEntitlementRepository = &RegistrationEntitlementRepositoryImpl{}

type RegistrationEntitlementRepository interface {
	// GetLatestRegistration(guildID string) (*models.RegistrationEntitlement, error)
	FindByQuery(q *models.RegistrationEntitlement) ([]models.RegistrationEntitlement, error)
	GetActive(q *models.RegistrationEntitlement) (*models.RegistrationEntitlement, error)
	Save(newRegistration *models.RegistrationEntitlement) error
}

type RegistrationEntitlementRepositoryImpl struct {
	DB *gorm.DB
}

func (r *RegistrationEntitlementRepositoryImpl) GetActive(q *models.RegistrationEntitlement) (*models.RegistrationEntitlement, error) {
	data := &models.RegistrationEntitlement{}
	actualQuery := *q
	actualQuery.UserID = nil
	actualQuery.Username = nil
	actualQuery.UsernameDiscriminator = nil
	userQuery := r.DB
	if q.UserID != nil {
		userQuery = userQuery.Or(&models.RegistrationEntitlement{UserID: q.UserID})
	}
	if q.Username != nil && q.UsernameDiscriminator != nil {
		userQuery = userQuery.Or(&models.RegistrationEntitlement{Username: q.Username, UsernameDiscriminator: q.UsernameDiscriminator})
	}
	if res := r.DB.Where(&actualQuery).Where(activeClause, time.Now()).Where(userQuery).Take(data); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return data, nil
}

func (r *RegistrationEntitlementRepositoryImpl) FindByQuery(q *models.RegistrationEntitlement) ([]models.RegistrationEntitlement, error) {
	lists := make([]models.RegistrationEntitlement, 0)
	// find only active ones
	if res := r.DB.Where(q).Where(activeClause, time.Now()).Find(&lists); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return lists, nil
		}
		return nil, res.Error
	}
	return lists, nil
}

func (r *RegistrationEntitlementRepositoryImpl) Save(newEntitlement *models.RegistrationEntitlement) error {
	if res := r.DB.Save(newEntitlement); res.Error != nil {
		return res.Error
	}
	return nil
}
