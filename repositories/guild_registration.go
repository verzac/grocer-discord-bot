package repositories

import (
	"time"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ GuildRegistrationRepository = &GuildRegistrationRepositoryImpl{}

type GuildRegistrationRepository interface {
	// GetLatestRegistration(guildID string) (*models.GuildRegistration, error)
	FindByQuery(q *models.GuildRegistration) ([]models.GuildRegistration, error)
	Count(q *models.GuildRegistration) (int64, error)
	Save(newRegistration *models.GuildRegistration) error
	Put(registrations ...models.GuildRegistration) error
}

type GuildRegistrationRepositoryImpl struct {
	DB *gorm.DB
}

func (r *GuildRegistrationRepositoryImpl) FindByQuery(q *models.GuildRegistration) ([]models.GuildRegistration, error) {
	lists := make([]models.GuildRegistration, 0)
	// find only active ones
	// okay this is actually bad because we're sending 3 separate SQL queries, so if anything gets too slow, this should be the one we optimise first
	if res := r.DB.Where(q).Where(activeClause, time.Now()).Preload("RegistrationEntitlement").Preload("RegistrationEntitlement.RegistrationTier").Find(&lists); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return lists, nil
		}
		return nil, res.Error
	}
	return lists, nil
}

func (r *GuildRegistrationRepositoryImpl) Save(newRegistration *models.GuildRegistration) error {
	if res := r.DB.Create(newRegistration); res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *GuildRegistrationRepositoryImpl) Put(registrations ...models.GuildRegistration) error {
	if res := r.DB.Save(registrations); res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *GuildRegistrationRepositoryImpl) Count(q *models.GuildRegistration) (out int64, err error) {
	if res := r.DB.Where(q).Where(activeClause, time.Now()).Count(&out); res.Error != nil {
		err = res.Error
	}
	return
}
