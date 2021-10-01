package repositories

import (
	"errors"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var (
	ErrGroceryListDuplicate = errors.New("A grocery list with that label already exists.")
	ErrGroceryListNotFound  = errors.New("Cannot find a grocery list with that label.")
)

var _ GroceryListRepository = &GroceryListRepositoryImpl{}

type GroceryListRepository interface {
	GetByQuery(q *models.GroceryList) (*models.GroceryList, error)
	FindByQuery(q *models.GroceryList) ([]models.GroceryList, error)
	Count(q *models.GroceryList) (existingCount int64, err error)
	CreateGroceryList(guildID string, listLabel string, fancyName string) (*models.GroceryList, error)
	Delete(groceryList *models.GroceryList) error
	Save(groceryList *models.GroceryList) error
}

type GroceryListRepositoryImpl struct {
	DB *gorm.DB
}

func (r *GroceryListRepositoryImpl) GetByQuery(q *models.GroceryList) (*models.GroceryList, error) {
	gl := models.GroceryList{}
	if res := r.DB.Where(q).Take(&gl); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return &gl, nil
}

func (r *GroceryListRepositoryImpl) FindByQuery(q *models.GroceryList) ([]models.GroceryList, error) {
	gLists := make([]models.GroceryList, 0)
	if res := r.DB.Where(q).Find(&gLists); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return gLists, nil
}

func (r *GroceryListRepositoryImpl) Count(q *models.GroceryList) (existingCount int64, err error) {
	countR := r.DB.Model(&models.GroceryList{}).Where(q).Count(&existingCount)
	if countR.Error != nil && countR.Error != gorm.ErrRecordNotFound {
		err = countR.Error
		return
	}
	return
}

func (r *GroceryListRepositoryImpl) CreateGroceryList(guildID string, listLabel string, fancyName string) (*models.GroceryList, error) {
	existingCount, err := r.Count(&models.GroceryList{GuildID: guildID, ListLabel: listLabel})
	if err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, ErrGroceryListDuplicate
	}

	newGroceryList := models.GroceryList{
		GuildID:   guildID,
		ListLabel: listLabel,
	}
	if fancyName == "" {
		newGroceryList.FancyName = nil
	} else {
		newGroceryList.FancyName = &fancyName
	}
	res := r.DB.Create(&newGroceryList)
	if res.Error != nil {
		return nil, res.Error
	}
	return &newGroceryList, nil
}

func (r *GroceryListRepositoryImpl) Delete(groceryList *models.GroceryList) error {
	res := r.DB.Delete(groceryList)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrGroceryListNotFound
	}
	return nil
}

func (r *GroceryListRepositoryImpl) Save(groceryList *models.GroceryList) error {
	return r.DB.Save(groceryList).Error
}
