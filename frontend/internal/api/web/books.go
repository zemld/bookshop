package web

import (
	"net/http"

	"bookshop/frontend/internal/domain/books/entities"
)

func (s Server) bookRoutes(m *http.ServeMux) {
	m.HandleFunc("GET /books", func(w http.ResponseWriter, r *http.Request) {
		items, err := s.Books.ListBooks(r.Context())
		if err != nil {
			fail(w, err)
			return
		}
		render(w, 200, page{Kind: booksPage, Title: "Книги", Books: items})
	})
	m.HandleFunc("GET /books/new", func(w http.ResponseWriter, r *http.Request) {
		publishers, err := s.Publishers.ListPublishers(r.Context())
		if err != nil {
			fail(w, err)
			return
		}
		render(w, 200, page{Kind: bookPage, Title: "Книга", Publishers: publishers, New: true})
	})
	m.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := routeID(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		form, err := s.LoadBookForm.LoadBookForm(r.Context(), id)
		if err != nil {
			fail(w, err)
			return
		}
		render(w, 200, page{Kind: bookPage, Title: "Книга", Publishers: form.Publishers, Book: form.Book})
	})
	m.HandleFunc("POST /books", s.saveBook)
	m.HandleFunc("POST /books/{id}", s.saveBook)
	m.HandleFunc("POST /books/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		id, err := routeID(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		if err := s.DeleteBook.DeleteBook(r.Context(), id); err != nil {
			fail(w, err)
			return
		}
		navigate(w, r, "/books")
	})
}

func (s Server) saveBook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fail(w, err)
		return
	}
	book, err := parseBookForm(r.Form)
	if err != nil {
		fail(w, err)
		return
	}
	var saved entities.Book
	if r.PathValue("id") == "" {
		saved, err = s.CreateBook.CreateBook(r.Context(), book)
	} else {
		id, parseErr := routeID(r.PathValue("id"))
		if parseErr != nil {
			fail(w, parseErr)
			return
		}
		saved, err = s.UpdateBook.UpdateBook(r.Context(), id, book)
	}
	if err != nil {
		fail(w, err)
		return
	}
	navigate(w, r, "/books/"+saved.ID.String())
}
