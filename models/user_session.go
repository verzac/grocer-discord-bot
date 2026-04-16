package models

import "time"

type UserSession struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	DiscordUserID      string    `gorm:"index;not null" json:"discord_user_id"`
	RefreshTokenHash   string    `gorm:"index;not null" json:"-"`
	RefreshTokenExpiry time.Time `gorm:"not null" json:"refresh_token_expiry"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
