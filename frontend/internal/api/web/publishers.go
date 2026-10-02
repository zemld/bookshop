package web

import (
	"net/http"

	"bookshop/frontend/internal/domain/publishers/entities"
)

func (s Server) publisherRoutes(m *http.ServeMux) {
	m.HandleFunc("GET /publishers", func(w http.ResponseWriter, r *http.Request) {
		items, err := s.Publishers.ListPublishers(r.Context())
		if err != nil {
			fail(w, err)
			return
		}
		render(w, 200, page{Kind: publishersPage, Title: "Издательства", Publishers: items})
	})
	m.HandleFunc("GET /publishers/new", func(w http.ResponseWriter, _ *http.Request) {
		render(w, 200, page{Kind: publisherPage, Title: "Издательство", New: true})
	})
	m.HandleFunc("GET /publishers/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := routeID(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		p, err := s.GetPublisher.GetPublisher(r.Context(), id)
		if err != nil {
			fail(w, err)
			return
		}
		render(w, 200, page{Kind: publisherPage, Title: "Издательство", Publisher: p})
	})
	m.HandleFunc("POST /publishers", s.savePublisher)
	m.HandleFunc("POST /publishers/{id}", s.savePublisher)
	m.HandleFunc("POST /publishers/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		id, err := routeID(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		if err := s.DeletePublisher.DeletePublisher(r.Context(), id); err != nil {
			fail(w, err)
			return
		}
		navigate(w, r, "/publishers")
	})
}

func (s Server) savePublisher(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fail(w, err)
		return
	}
	p := entities.Publisher{Name: r.Form.Get("name")}
	var saved entities.Publisher
	var err error
	if r.PathValue("id") == "" {
		saved, err = s.CreatePublisher.CreatePublisher(r.Context(), p)
	} else {
		var id = r.PathValue("id")
		parsed, parseErr := routeID(id)
		if parseErr != nil {
			fail(w, parseErr)
			return
		}
		saved, err = s.UpdatePublisher.UpdatePublisher(r.Context(), parsed, p)
	}
	if err != nil {
		fail(w, err)
		return
	}
	navigate(w, r, "/publishers/"+saved.ID.String())
}
