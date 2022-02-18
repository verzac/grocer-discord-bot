package utils

import "github.com/go-playground/validator/v10"

// CustomValidator used to implement top-level request validation logic
type CustomValidator struct {
	validator *validator.Validate
}

// Validate implements Echo's validation interface
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// NewCustomValidator returns a new CustomValidator instance
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}
