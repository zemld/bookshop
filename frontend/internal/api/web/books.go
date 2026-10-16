package web

import (
	"net/http"

	"bookshop/frontend/internal/domain/books/entities"
	"bookshop/frontend/internal/requestlog"
)

func (s Server) registerBookRoutes(m *http.ServeMux) {
	m.HandleFunc("GET /books", func(w http.ResponseWriter, r *http.Request) {
		items, err := s.Books.ListBooks(r.Context())
		if err != nil {
			s.renderErrorPage(w, r, err)
			return
		}
		renderPage(w, 200, page{Kind: booksPage, Title: "Книги", Books: items})
	})
	m.HandleFunc("GET /books/new", func(w http.ResponseWriter, r *http.Request) {
		publishers, err := s.Publishers.ListPublishers(r.Context())
		if err != nil {
			s.renderErrorPage(w, r, err)
			return
		}
		renderPage(w, 200, page{Kind: bookPage, Title: "Книга", Publishers: publishers, New: true})
	})
	m.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseRouteID(r.PathValue("id"))
		if err != nil {
			s.renderErrorPage(w, r, err)
			return
		}
		form, err := s.LoadBookForm.LoadBookForm(r.Context(), id)
		if err != nil {
			s.renderErrorPage(w, r, err)
			return
		}
		renderPage(w, 200, page{Kind: bookPage, Title: "Книга", Publishers: form.Publishers, Book: form.Book})
	})
	m.HandleFunc("POST /books", s.saveBook)
	m.HandleFunc("POST /books/{id}", s.saveBook)
	m.HandleFunc("POST /books/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseRouteID(r.PathValue("id"))
		if err != nil {
			s.renderErrorPage(w, r, err)
			return
		}
		if err := s.DeleteBook.DeleteBook(r.Context(), id); err != nil {
			s.renderErrorPage(w, r, err)
			return
		}
		navigateToPage(w, r, "/books")
	})
}

func (s Server) saveBook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderErrorPage(w, r, err)
		return
	}
	book, err := parseBookForm(r.PostForm)
	if err != nil {
		s.renderBookFormError(w, r, book, err)
		return
	}
	var saved entities.Book
	if r.PathValue("id") == "" {
		saved, err = s.CreateBook.CreateBook(r.Context(), book)
	} else {
		id, parseErr := parseRouteID(r.PathValue("id"))
		if parseErr != nil {
			s.renderErrorPage(w, r, parseErr)
			return
		}
		saved, err = s.UpdateBook.UpdateBook(r.Context(), id, book)
	}
	if err != nil {
		s.renderBookFormError(w, r, book, err)
		return
	}
	navigateToPage(w, r, "/books/"+saved.ID.String())
}

func (s Server) renderBookFormError(w http.ResponseWriter, r *http.Request, book entities.Book, err error) {
	if r.PathValue("id") != "" {
		id, parseErr := parseRouteID(r.PathValue("id"))
		if parseErr != nil {
			s.renderErrorPage(w, r, parseErr)
			return
		}
		book.ID = id
	}
	data := page{Kind: bookPage, Title: "Книга", Book: book,
		New: r.PathValue("id") == "", Form: r.PostForm}
	publishers, loadErr := s.Publishers.ListPublishers(r.Context())
	if loadErr != nil {
		if s.Logger != nil {
			s.Logger.ErrorContext(r.Context(), "failed to reload publishers for book form",
				append(append([]any{"event", "book_form.publishers_load_failed"}, requestlog.BuildAttributes(r.Context())...), "error", loadErr)...)
		}
		logUnexpectedError(r.Context(), s.Logger, err)
		code, message, fields := classifyPublicError(err)
		data.Error = message + " Не удалось загрузить издательства. Повторите попытку позже."
		data.Fields = fields
		renderPage(w, code, data)
		return
	}
	data.Publishers = publishers
	s.renderFormError(w, r, err, data)
}
