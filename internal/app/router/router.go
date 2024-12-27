package router

import (
	"github.com/DeneesK/short-url/internal/app/handlers"
	"github.com/DeneesK/short-url/internal/app/middlewares"
	"github.com/go-chi/chi/v5"
)

type Repository interface {
	SaveURL(string) (string, error)
	GetURL(string) (string, error)
}

func NewRouter(rep Repository, memUsageLimit float64, memCheckingType string) *chi.Mux {
	r := chi.NewRouter()

	m := middlewares.NewMemoryControlMiddleware(memUsageLimit, memCheckingType)
	r.With(m).Post("/", handlers.URLSaver(rep))

	r.Get("/{id}", handlers.URLRedirect(rep))

	return r
}
