package models

import "time"

type GrohereRecord struct {
	ID               uint `gorm:"primaryKey"`
	GuildID          string
	GroceryListID    *uint
	GroceryList      *GroceryList
	GrohereChannelID string
	GrohereMessageID string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	// LastSeenAt       *time.Time
}
