package web

import (
	"bytes"
	"context"
	"net/http"

	bookentities "bookshop/frontend/internal/domain/books/entities"
	publisherentities "bookshop/frontend/internal/domain/publishers/entities"
)

type page struct {
	Kind       pageKind
	Title      string
	Error      string
	Publishers []publisherentities.Publisher
	Books      []bookentities.Book
	Publisher  publisherentities.Publisher
	Book       bookentities.Book
	New        bool
}

type pageKind uint8

const (
	publishersPage pageKind = iota + 1
	booksPage
	publisherPage
	bookPage
	errorPage
)

func render(w http.ResponseWriter, code int, data page) {
	switch data.Kind {
	case publishersPage, booksPage, publisherPage, bookPage, errorPage:
	default:
		http.Error(w, "unsupported page", http.StatusInternalServerError)
		return
	}
	var output bytes.Buffer
	if err := document(data).Render(context.Background(), &output); err != nil {
		http.Error(w, "render page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, _ = output.WriteTo(w)
}

func fail(w http.ResponseWriter, err error) {
	render(w, http.StatusBadRequest, page{Kind: errorPage, Title: "Ошибка", Error: err.Error()})
}
