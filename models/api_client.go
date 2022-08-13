package models

import (
	"gorm.io/gorm"
)

type ApiClient struct {
	gorm.Model
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedByID string         `json:"created_by"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
	// GuildID      string    `gorm:"index;not null" json:"guild_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"-"` // never return this as part of the struct ya wanker
	// Scope is a custom scope - you can chuck anything you want here to indicate what the API key is used for. Generally you put in the format guild:123456789
	Scope string `json:"scope"`
}
