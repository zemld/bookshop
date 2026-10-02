package publishers

import publisherports "bookshop/frontend/internal/ports/publishers"

type Service struct {
	Publishers publisherports.WritePublisher
}
