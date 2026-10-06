package shared

import "errors"

var (
	ErrInvalid  = errors.New("invalid_input")
	ErrConflict = errors.New("conflict")
	ErrNotFound = errors.New("not_found")
)
