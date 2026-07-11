package demo

import "errors"

var (
	ErrInvalidPassword = errors.New("invalid demo password")
	ErrNotSeeded       = errors.New("demo account not configured; seed it by logging in with the demo Discord account")
	ErrInternal        = errors.New("cannot issue demo session")
)
