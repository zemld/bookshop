package web

import (
	"net/http"

	"bookshop/frontend/internal/domain/publishers/entities"
)

func (s Server) registerPublisherRoutes(m *http.ServeMux) {
	m.HandleFunc("GET /publishers", func(w http.ResponseWriter, r *http.Request) {
		items, err := s.Publishers.ListPublishers(r.Context())
		if err != nil {
			renderErrorPage(w, err)
			return
		}
		renderPage(w, 200, page{Kind: publishersPage, Title: "Издательства", Publishers: items})
	})
	m.HandleFunc("GET /publishers/new", func(w http.ResponseWriter, _ *http.Request) {
		renderPage(w, 200, page{Kind: publisherPage, Title: "Издательство", New: true})
	})
	m.HandleFunc("GET /publishers/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseRouteID(r.PathValue("id"))
		if err != nil {
			renderErrorPage(w, err)
			return
		}
		p, err := s.GetPublisher.GetPublisher(r.Context(), id)
		if err != nil {
			renderErrorPage(w, err)
			return
		}
		renderPage(w, 200, page{Kind: publisherPage, Title: "Издательство", Publisher: p})
	})
	m.HandleFunc("POST /publishers", s.savePublisher)
	m.HandleFunc("POST /publishers/{id}", s.savePublisher)
	m.HandleFunc("POST /publishers/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseRouteID(r.PathValue("id"))
		if err != nil {
			renderErrorPage(w, err)
			return
		}
		if err := s.DeletePublisher.DeletePublisher(r.Context(), id); err != nil {
			renderErrorPage(w, err)
			return
		}
		navigateToPage(w, r, "/publishers")
	})
}

func (s Server) savePublisher(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderErrorPage(w, err)
		return
	}
	p := entities.Publisher{Name: r.PostForm.Get("name")}
	var saved entities.Publisher
	var err error
	if r.PathValue("id") == "" {
		saved, err = s.CreatePublisher.CreatePublisher(r.Context(), p)
	} else {
		var id = r.PathValue("id")
		parsed, parseErr := parseRouteID(id)
		if parseErr != nil {
			renderErrorPage(w, parseErr)
			return
		}
		saved, err = s.UpdatePublisher.UpdatePublisher(r.Context(), parsed, p)
		p.ID = parsed
	}
	if err != nil {
		renderFormError(w, err, page{Kind: publisherPage, Title: "Издательство",
			Publisher: p, New: r.PathValue("id") == "", Form: r.PostForm})
		return
	}
	navigateToPage(w, r, "/publishers/"+saved.ID.String())
}
