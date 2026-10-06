package web

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	bookentities "bookshop/frontend/internal/domain/books/entities"
	publisherentities "bookshop/frontend/internal/domain/publishers/entities"
)

type page struct {
	Kind       pageKind
	Title      string
	Error      string
	Fields     map[string]string
	Form       url.Values
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

func renderPage(w http.ResponseWriter, code int, data page) {
	switch data.Kind {
	case publishersPage, booksPage, publisherPage, bookPage, errorPage:
	default:
		http.Error(w, "unsupported page", http.StatusInternalServerError)
		return
	}
	var output bytes.Buffer
	if err := renderDocument(data).Render(context.Background(), &output); err != nil {
		http.Error(w, "render page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, _ = output.WriteTo(w)
}

func renderErrorPage(w http.ResponseWriter, err error) {
	code, message, fields := classifyPublicError(err)
	renderPage(w, code, page{Kind: errorPage, Title: "Ошибка", Error: message, Fields: fields})
}
