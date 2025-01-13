package router

import (
	"github.com/DeneesK/short-url/internal/app/handlers"
	"github.com/go-chi/chi/v5"
)

type Repository interface {
	SaveURL(string) (string, error)
	GetURL(string) (string, error)
}

func NewRouter(rep Repository) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/", handlers.URLSaver(rep))
	r.Get("/{id}", handlers.URLRedirect(rep))

	return r
}
