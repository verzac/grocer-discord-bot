package models

import (
	"fmt"
	"time"

	"github.com/andanhm/go-prettytime"
)

type GroceryEntry struct {
	ID          uint `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ItemDesc    string `gorm:"not null"`
	GuildID     string `gorm:"index;not null"`
	UpdatedByID *string
}

func (g *GroceryEntry) GetUpdatedByString() string {
	updatedByString := ""
	if g.UpdatedByID != nil {
		updatedByString = fmt.Sprintf("(updated by <@%s> %s)", *g.UpdatedByID, prettytime.Format(g.UpdatedAt))
	}
	return updatedByString
}
