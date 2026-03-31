package models

import "time"

type UserSession struct {
	DiscordUserID         string    `gorm:"primaryKey;column:discord_user_id" json:"discord_user_id"`
	DiscordAccessTokenEnc  []byte    `gorm:"column:discord_access_token_enc;not null" json:"-"`
	DiscordRefreshTokenEnc []byte    `gorm:"column:discord_refresh_token_enc;not null" json:"-"`
	DiscordTokenExpiresAt  *time.Time `gorm:"column:discord_token_expires_at" json:"-"`
	GroRefreshTokenHash    string    `gorm:"column:gro_refresh_token_hash;not null" json:"-"`
	GroRefreshExpiresAt    time.Time `gorm:"column:gro_refresh_expires_at;not null" json:"-"`
	CreatedAt              time.Time `gorm:"not null" json:"-"`
	UpdatedAt              time.Time `gorm:"not null" json:"-"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}
