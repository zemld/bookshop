package convert

import (
	"bookshop/backend/internal/api/rest/ogen"
	bookentities "bookshop/backend/internal/domain/books/entities"
	publisherentities "bookshop/backend/internal/domain/publishers/entities"
)

func ToOgenPublisher(p publisherentities.Publisher) ogen.Publisher {
	return ogen.Publisher{ID: p.ID, Name: p.Name}
}

func ToOgenBook(b bookentities.Book) ogen.Book {
	return ogen.Book{ID: b.ID, Author: b.Author, Name: b.Name, Year: int(b.Year),
		Price: b.Price, PublisherId: b.PublisherID, PublicationYear: int(b.PublicationYear), Quantity: b.Quantity}
}

func ToDomainBook(b *ogen.BookInput) bookentities.Book {
	return bookentities.Book{Author: b.Author, Name: b.Name, Year: bookentities.Year(b.Year),
		Price: b.Price, PublisherID: b.PublisherId, PublicationYear: bookentities.Year(b.PublicationYear), Quantity: b.Quantity}
}
