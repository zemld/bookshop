package web

import (
	"net/http"

	"bookshop/frontend/internal/domain/books"
	"bookshop/frontend/internal/domain/publishers"
	bookports "bookshop/frontend/internal/ports/books"
	publisherports "bookshop/frontend/internal/ports/publishers"
)

// Direct ports are used for unadorned list/get/delete transport operations;
// services exist only where the UI adds normalization or coordinates calls.
type Server struct {
	Books           bookports.ListBooks
	DeleteBook      bookports.DeleteBook
	Publishers      publisherports.ListPublishers
	GetPublisher    publisherports.GetPublisher
	DeletePublisher publisherports.DeletePublisher
	LoadBookForm    books.LoadBookForm
	CreateBook      bookports.CreateBook
	UpdateBook      bookports.UpdateBook
	CreatePublisher publishers.CreatePublisher
	UpdatePublisher publishers.UpdatePublisher
}

func (s Server) Routes() http.Handler {
	m := http.NewServeMux()
	m.Handle("GET /assets/", http.FileServer(http.FS(assets)))
	m.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/books", http.StatusSeeOther)
	})
	s.publisherRoutes(m)
	s.bookRoutes(m)
	return m
}
