package books

import (
	"bookshop/backend/internal/domain/books"
	bookrepository "bookshop/backend/internal/ports/books"
)

type Service struct{ Repository bookrepository.Repository }

var _ books.Books = Service{}
