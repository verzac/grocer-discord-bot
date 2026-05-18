package models

import "time"

type WaitlistIos struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DiscordUserID string    `gorm:"index;not null" json:"discord_user_id"`
	Email         string    `gorm:"not null" json:"email"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
}
