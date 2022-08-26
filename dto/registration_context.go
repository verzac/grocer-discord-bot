package dto

type RegistrationContext struct {
	MaxGroceryListsPerServer   int
	MaxGroceryEntriesPerServer int
	IsDefault                  bool
	RegistrationsOwnersMention []string
}
