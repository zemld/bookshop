package web

import (
	"net/http"

	"bookshop/frontend/internal/domain/books"
	"bookshop/frontend/internal/domain/publishers"
	bookports "bookshop/frontend/internal/ports/books"
	publisherports "bookshop/frontend/internal/ports/publishers"
)

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

func (s Server) CreateHandler() http.Handler {
	m := http.NewServeMux()
	m.Handle("GET /assets/", http.FileServer(http.FS(assets)))
	m.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/books", http.StatusSeeOther)
	})
	s.registerPublisherRoutes(m)
	s.registerBookRoutes(m)
	return m
}
