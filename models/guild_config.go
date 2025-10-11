package models

import "time"

type GuildConfig struct {
	GuildID                 string `gorm:"primaryKey"`
	GrohereChannelID        *string
	GrohereMessageID        *string
	CreatedAt               time.Time
	UpdatedAt               time.Time
	UseEphemeral            bool
	UseGrobulkAppend        bool // legacy opt-in flag for backwards compatibility - most guilds should have this be disabled
	LastAnnouncementVersion int
	// LastSeenAt       *time.Time
}
