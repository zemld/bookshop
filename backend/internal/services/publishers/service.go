package publishers

import (
	"bookshop/backend/internal/domain/publishers"
	publisherrepository "bookshop/backend/internal/ports/publishers"
)

type Service struct {
	Repository publisherrepository.Repository
}

var _ publishers.Publishers = Service{}
