package books

import (
	bookports "bookshop/frontend/internal/ports/books"
	publisherports "bookshop/frontend/internal/ports/publishers"
)

type Service struct {
	Books      bookports.GetBook
	Publishers publisherports.ListPublishers
}
