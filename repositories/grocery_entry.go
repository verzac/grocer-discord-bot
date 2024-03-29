package repositories

import (
	"context"
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var _ GroceryEntryRepository = &GroceryEntryRepositoryImpl{}

const (
	queryGroceryListIDIsNil = "grocery_list_id IS NULL"
)

var (
	ErrGroceryListGuildIDMismatch = &RepositoryError{
		ErrCode: ErrInternal,
		Message: "Grocery list's guildID and passed in guildID does not match.",
	}
)

type GroceryEntryQueryOpts struct {
	IsStrongNilForGroceryListID bool
}

type GroceryEntryRepository interface {
	WithContext(ctx context.Context) GroceryEntryRepository
	GetByItemIndex(q *models.GroceryEntry, itemIndex int) (*models.GroceryEntry, error)
	FindByQuery(q *models.GroceryEntry) ([]models.GroceryEntry, error)
	FindByQueryWithConfig(q *models.GroceryEntry, config GroceryEntryQueryOpts) ([]models.GroceryEntry, error)
	AddToGroceryList(groceryList *models.GroceryList, groceryEntries []models.GroceryEntry, guildID string) *RepositoryError
	ClearGroceryList(groceryList *models.GroceryList, guildID string) (rowsAffected int64, err *RepositoryError)
	Put(g *models.GroceryEntry) error
	GetCount(query *models.GroceryEntry) (count int64, err error)
}

type GroceryEntryRepositoryImpl struct {
	DB *gorm.DB
}

func (r *GroceryEntryRepositoryImpl) WithContext(ctx context.Context) GroceryEntryRepository {
	return &GroceryEntryRepositoryImpl{
		DB: r.DB.WithContext(ctx),
	}
}

func (r *GroceryEntryRepositoryImpl) GetByItemIndex(q *models.GroceryEntry, itemIndex int) (*models.GroceryEntry, error) {
	g := models.GroceryEntry{}
	dbQuery := r.DB.Where(q)
	if q.GroceryListID == nil {
		// force, since itemIndex is only relevant for a particular grocery list
		dbQuery = dbQuery.Where(queryGroceryListIDIsNil)
	}
	if res := dbQuery.Offset(itemIndex - 1).First(&g); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return &g, nil
}

func (r *GroceryEntryRepositoryImpl) FindByQueryWithConfig(q *models.GroceryEntry, config GroceryEntryQueryOpts) ([]models.GroceryEntry, error) {
	entries := make([]models.GroceryEntry, 0)
	dbQuery := r.DB.Where(q)
	if config.IsStrongNilForGroceryListID && q.GroceryListID == nil {
		dbQuery = dbQuery.Where(queryGroceryListIDIsNil)
	}
	if res := dbQuery.Find(&entries); res.Error != nil {
		if res.Error != gorm.ErrRecordNotFound {
			return nil, res.Error
		}
	}
	return entries, nil
}

func (r *GroceryEntryRepositoryImpl) FindByQuery(q *models.GroceryEntry) ([]models.GroceryEntry, error) {
	return r.FindByQueryWithConfig(q, GroceryEntryQueryOpts{})
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
	if err := r.Create(groceryEntries); err != nil {
		return err
	}
	return nil
}

func (r *GroceryEntryRepositoryImpl) Create(groceryEntries []models.GroceryEntry) *RepositoryError {
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

func (r *GroceryEntryRepositoryImpl) Put(g *models.GroceryEntry) error {
	if r := r.DB.Save(g); r.Error != nil {
		return r.Error
	}
	return nil
}

func (r *GroceryEntryRepositoryImpl) GetCount(query *models.GroceryEntry) (count int64, err error) {
	if r := r.DB.Model(&models.GroceryEntry{}).Where(query).Count(&count); r.Error != nil {
		return 0, r.Error
	}
	return count, nil
}
