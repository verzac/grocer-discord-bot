package models

type RegistrationTier struct {
	ID              int
	Name            string
	MaxGroceryEntry *int
	MaxGroceryList  *int
}
