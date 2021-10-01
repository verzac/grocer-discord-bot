package utils

import (
	"errors"
	"regexp"
)

const sublistFormat = `^[a-zA-Z]+$`

var (
	ErrInvalidSublistLabelFmt = errors.New("Grocery list labels are case-insensitive and must only contain letters.")
)

func ValidateSublistLabel(label string) error {
	if !regexp.MustCompile(sublistFormat).MatchString(label) {
		return ErrInvalidSublistLabelFmt
	}
	return nil
}
