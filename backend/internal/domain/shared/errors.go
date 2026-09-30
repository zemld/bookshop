package shared

import "errors"

var (
	ErrInvalid  = errors.New("invalid input")
	ErrConflict = errors.New("duplicate or referenced record")
	ErrNotFound = errors.New("record not found")
)
