package models

import "time"

type RegistrationEntitlement struct {
	UserID             string `gorm:"primaryKey"`
	ExpiresAt          *time.Time
	ExternalID         *string
	ExternalIDType     *string
	MaxRedemption      int
	RegistrationTierID int
	RegistrationTier   *RegistrationTier
}
