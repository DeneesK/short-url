package router

import (
	"github.com/go-chi/chi/v5"
)

type URLService interface {
	ShortenURL(string) (string, error)
	FindByShortened(string) (string, error)
}

func NewRouter(service URLService) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/", ShortenURL(service))
	r.Get("/{id}", URLRedirect(service))

	return r
}
