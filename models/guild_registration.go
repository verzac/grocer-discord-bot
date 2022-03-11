package models

import "time"

type GuildRegistration struct {
	ID                            int
	GuildID                       string
	RegistrationEntitlementUserID string
	RegistrationEntitlement       *RegistrationEntitlement
	CreatedAt                     time.Time
	ExpiresAt                     *time.Time
}
