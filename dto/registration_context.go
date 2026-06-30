package dto

type RegistrationContext struct {
	// MaxGroceryListsPerServer includes the implicit default list. Stored GroceryList rows only count named lists.
	MaxGroceryListsPerServer   int
	MaxGroceryEntriesPerServer int
	IsDefault                  bool
	RegistrationsOwnersMention []string
}
