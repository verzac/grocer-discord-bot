package models

import "time"

type RegistrationEntitlement struct {
	ID                    uint `gorm:"primaryKey"`
	UserID                *string
	Username              *string
	UsernameDiscriminator *string
	ExpiresAt             *time.Time
	ExternalID            *string
	ExternalIDType        *string
	MaxRedemption         int
	RegistrationTierID    int
	RegistrationTier      *RegistrationTier
}
