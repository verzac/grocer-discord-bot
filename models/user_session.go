package models

import (
	"time"

	"gorm.io/gorm"
)

type UserSession struct {
	gorm.Model
	DiscordUserID       string    `gorm:"uniqueIndex;not null" json:"discord_user_id"`
	DiscordAccessToken  string    `gorm:"not null" json:"-"`
	DiscordRefreshToken string    `gorm:"not null" json:"-"`
	DiscordTokenExpiry  time.Time `gorm:"not null" json:"-"`
	RefreshTokenHash    string    `gorm:"index;not null" json:"-"`
	RefreshTokenExpiry  time.Time `gorm:"not null" json:"-"`
}
