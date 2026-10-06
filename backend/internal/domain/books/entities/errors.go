package entities

import (
	"fmt"

	"bookshop/backend/internal/domain/shared"
)

var (
	ErrAuthorRequired            = fmt.Errorf("book_author_required: %w", shared.ErrInvalid)
	ErrNameRequired              = fmt.Errorf("book_name_required: %w", shared.ErrInvalid)
	ErrYearInvalid               = fmt.Errorf("book_year_invalid: %w", shared.ErrInvalid)
	ErrPriceNegative             = fmt.Errorf("book_price_negative: %w", shared.ErrInvalid)
	ErrPublisherRequired         = fmt.Errorf("book_publisher_required: %w", shared.ErrInvalid)
	ErrPublicationYearInvalid    = fmt.Errorf("book_publication_year_invalid: %w", shared.ErrInvalid)
	ErrPublicationYearBeforeYear = fmt.Errorf("book_publication_year_before_year: %w", shared.ErrInvalid)
	ErrQuantityNegative          = fmt.Errorf("book_quantity_negative: %w", shared.ErrInvalid)
	ErrDuplicate                 = fmt.Errorf("book_duplicate: %w", shared.ErrConflict)
	ErrPublisherNotFound         = fmt.Errorf("book_publisher_not_found: %w", shared.ErrConflict)
)
