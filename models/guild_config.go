package models

import "time"

type GuildConfig struct {
	GuildID          string `gorm:"primaryKey"`
	GrohereChannelID *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
