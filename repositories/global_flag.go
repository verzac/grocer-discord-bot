package repositories

import (
	"errors"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ GlobalFlagRepository = &GlobalFlagRepositoryImpl{}

type GlobalFlagRepository interface {
	GetFlag(key string) (string, error)
	SetFlag(key, value string) error
}

type GlobalFlagRepositoryImpl struct {
	DB *gorm.DB
}

func (r *GlobalFlagRepositoryImpl) GetFlag(key string) (string, error) {
	var flag models.GlobalFlag
	if res := r.DB.Where("key = ?", key).First(&flag); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", res.Error
	}
	return flag.Value, nil
}

func (r *GlobalFlagRepositoryImpl) SetFlag(key, value string) error {
	flag := models.GlobalFlag{
		Key:   key,
		Value: value,
	}
	if res := r.DB.Save(&flag); res.Error != nil {
		return res.Error
	}
	return nil
}
