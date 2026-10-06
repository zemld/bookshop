package web

import (
	"strconv"

	"github.com/a-h/templ"
)

func buildBookFormAction(p page) templ.SafeURL {
	if p.New {
		return "/books"
	}
	return templ.SafeURL("/books/" + p.Book.ID.String())
}

func buildPublisherFormAction(p page) templ.SafeURL {
	if p.New {
		return "/publishers"
	}
	return templ.SafeURL("/publishers/" + p.Publisher.ID.String())
}

func formatBookNumber(newBook bool, value int64) string {
	if newBook {
		return ""
	}
	return strconv.FormatInt(value, 10)
}

func getBookInputValue(p page, field string) string {
	if p.Form != nil {
		return p.Form.Get(field)
	}
	switch field {
	case "author":
		return p.Book.Author
	case "name":
		return p.Book.Name
	case "year":
		return formatBookNumber(p.New, int64(p.Book.Year))
	case "price":
		return formatBookNumber(p.New, p.Book.Price)
	case "publicationYear":
		return formatBookNumber(p.New, int64(p.Book.PublicationYear))
	case "quantity":
		return formatBookNumber(p.New, p.Book.Quantity)
	case "publisherId":
		return p.Book.PublisherID.String()
	default:
		return ""
	}
}

func getPublisherInputValue(p page) string {
	if p.Form != nil {
		return p.Form.Get("name")
	}
	return p.Publisher.Name
}

func getErrorDescriptionID(p page, field string) string {
	if p.Fields[field] != "" {
		return field + "-error"
	}
	return ""
}
