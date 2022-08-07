package models

import (
	"time"
)

type ApiKey struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedByID string    `json:"created_by"`
	// GuildID      string    `gorm:"index;not null" json:"guild_id"`
	ApiKeyHashed string `json:"-"` // never return this as part of the struct
	// Scope is a custom scope - you can chuck anything you want here to indicate what the API key is used for. Generally you put in the format guild:123456789
	Scope string `json:"scope"`
}
