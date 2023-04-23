package models

import "time"

type GuildConfig struct {
	GuildID          string `gorm:"primaryKey"`
	GrohereChannelID *string
	GrohereMessageID *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	UseEphemeral     bool
	// LastSeenAt       *time.Time
}
