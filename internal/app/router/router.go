package router

import (
	"time"

	"github.com/DeneesK/short-url/internal/app/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

type Repository interface {
	SaveURL(string) (string, error)
	GetURL(string) (string, error)
}

func NewRouter(rep Repository, rateLimit int) *chi.Mux {
	r := chi.NewRouter()

	r.Use(httprate.LimitByIP(rateLimit, time.Minute))

	r.Post("/", handlers.URLSaver(rep))
	r.Get("/{id}", handlers.URLRedirect(rep))

	return r
}
