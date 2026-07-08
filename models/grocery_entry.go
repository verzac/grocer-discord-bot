package models

import (
	"time"
)

type GroceryEntry struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	ItemDesc      string       `gorm:"not null" json:"item_desc" validate:"required"`
	GuildID       string       `gorm:"index;not null" json:"guild_id"`
	UpdatedByID   *string      `json:"updated_by_id"`
	GroceryListID *uint        `json:"grocery_list_id"`
	GroceryList   *GroceryList `json:"grocery_list"`
}
