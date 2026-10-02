package web

import (
	"strconv"

	"github.com/a-h/templ"
)

func bookAction(p page) templ.SafeURL {
	if p.New {
		return "/books"
	}
	return templ.SafeURL("/books/" + p.Book.ID.String())
}

func publisherAction(p page) templ.SafeURL {
	if p.New {
		return "/publishers"
	}
	return templ.SafeURL("/publishers/" + p.Publisher.ID.String())
}

func bookNumber(newBook bool, value int64) string {
	if newBook {
		return ""
	}
	return strconv.FormatInt(value, 10)
}
