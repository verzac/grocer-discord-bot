package models

import "time"

type GuildRegistration struct {
	ID                        int
	GuildID                   string
	RegistrationEntitlementID uint
	RegistrationEntitlement   *RegistrationEntitlement
	CreatedAt                 time.Time
	ExpiresAt                 *time.Time
}
