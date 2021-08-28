package repositories

import (
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ GroceryEntryRepository = &GroceryEntryRepositoryImpl{}

const groceryEntryLimit = 100

var (
	ErrGroceryListGuildIDMismatch = &RepositoryError{
		ErrCode: ErrInternal,
		Message: "Grocery list's guildID and passed in guildID does not match.",
	}
)

type GroceryEntryRepository interface {
	GetByQuery(q *models.GroceryEntry) (*models.GroceryEntry, error)
	FindByQuery(q *models.GroceryEntry) ([]models.GroceryEntry, error)
	// FindByGroceryList(groceryList *models.GroceryList) ([]models.GroceryEntry, error)
	AddToGroceryList(groceryList *models.GroceryList, groceryEntries []models.GroceryEntry, guildID string) *RepositoryError
	ClearGroceryList(groceryList *models.GroceryList, guildID string) (rowsAffected int64, err *RepositoryError)
}

type GroceryEntryRepositoryImpl struct {
	DB *gorm.DB
}

func (r *GroceryEntryRepositoryImpl) GetByQuery(q *models.GroceryEntry) (*models.GroceryEntry, error) {
	g := models.GroceryEntry{}
	if res := r.DB.Where(q).Take(g); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return &g, nil
}

func (r *GroceryEntryRepositoryImpl) FindByQuery(q *models.GroceryEntry) ([]models.GroceryEntry, error) {
	entries := make([]models.GroceryEntry, 0)
	if res := r.DB.Where(q).Find(&entries); res.Error != nil {
		if res.Error != gorm.ErrRecordNotFound {
			return nil, res.Error
		}
	}
	return entries, nil
}

func (r *GroceryEntryRepositoryImpl) FindByGroceryList(groceryList *models.GroceryList) ([]models.GroceryEntry, error) {
	return r.FindByQuery(&models.GroceryEntry{GroceryListID: &groceryList.ID})
}

func (r *GroceryEntryRepositoryImpl) AddToGroceryList(groceryList *models.GroceryList, groceryEntries []models.GroceryEntry, guildID string) *RepositoryError {
	// validate
	if groceryList != nil && groceryList.GuildID != guildID {
		return ErrGroceryListGuildIDMismatch
	}
	for i := range groceryEntries {
		if groceryEntries[i].ID > 0 {
			return &RepositoryError{
				ErrCode: ErrCodeValidationError,
				Message: fmt.Sprintf(
					"Whoops, *%s* seems to exist already... This shouldn't happen; could you please get in touch with my maintainer? Thank you!",
					groceryEntries[i].ItemDesc,
				),
			}
		}
		groceryEntries[i].GuildID = guildID
		if groceryList != nil {
			groceryEntries[i].GroceryListID = &groceryList.ID
		}
	}
	if err := r.checkLimit(guildID, len(groceryEntries)); err != nil {
		return err
	}
	if res := r.DB.Omit("GroceryList").Create(&groceryEntries); res.Error != nil {
		return &RepositoryError{
			ErrCode: ErrInternal,
			Message: res.Error.Error(),
		}
	}
	return nil
}

func (r *GroceryEntryRepositoryImpl) ClearGroceryList(groceryList *models.GroceryList, guildID string) (rowsAffected int64, err *RepositoryError) {
	// validate
	if groceryList != nil && groceryList.GuildID != guildID {
		return 0, ErrGroceryListGuildIDMismatch
	}
	var res *gorm.DB
	if groceryList != nil {
		res = r.DB.Delete(models.GroceryEntry{}, "guild_id = ? AND grocery_list_id = ?", guildID, groceryList.ID)
	} else {
		res = r.DB.Delete(models.GroceryEntry{}, "guild_id = ? AND grocery_list_id IS NULL", guildID)
	}
	if res.Error != nil {
		return 0, &RepositoryError{
			Message: res.Error.Error(),
			ErrCode: ErrInternal,
		}
	}
	return res.RowsAffected, nil
}

func (r *GroceryEntryRepositoryImpl) checkLimit(guildID string, newItemCount int) *RepositoryError {
	var count int64
	if r := r.DB.Model(&models.GroceryEntry{}).Where(&models.GroceryEntry{GuildID: guildID}).Count(&count); r.Error != nil {
		return &RepositoryError{
			ErrCode: ErrInternal,
			Message: r.Error.Error(),
		}
	}
	if count+int64(newItemCount) > groceryEntryLimit {
		return &RepositoryError{
			ErrCode: ErrCodeValidationError,
			Message: fmt.Sprintf("Whoops, you've gone over the limit allowed by the bot (max %d grocery entries per server). Please log an issue through our Discord server (look at `!grohelp`) to request an increase! Thank you for being a power user! :tada:", groceryEntryLimit),
		}
	}
	return nil
}
