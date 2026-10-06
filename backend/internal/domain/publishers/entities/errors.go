package entities

import (
	"fmt"

	"bookshop/backend/internal/domain/shared"
)

var (
	ErrNameRequired  = fmt.Errorf("publisher_name_required: %w", shared.ErrInvalid)
	ErrNameDuplicate = fmt.Errorf("publisher_name_duplicate: %w", shared.ErrConflict)
	ErrReferenced    = fmt.Errorf("publisher_referenced: %w", shared.ErrConflict)
)
