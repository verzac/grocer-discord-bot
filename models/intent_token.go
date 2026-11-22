package models

import "time"

type IntentToken struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	DiscordUserID string     `gorm:"index;not null" json:"discord_user_id"`
	TokenHash     string     `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt     time.Time  `gorm:"not null" json:"expires_at"`
	LastUsedAt    *time.Time `json:"last_used_at"`
	CreatedAt     time.Time  `json:"created_at"`
}
