package router

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type URLService interface {
	ShortenURL(string) (string, error)
	FindByShortened(string) (string, error)
}

func NewRouter(service URLService, log *zap.SugaredLogger) *chi.Mux {
	r := chi.NewRouter()

	loggingMiddleware := NewLoggingMiddleware(log)
	r.Use(loggingMiddleware)

	r.Post("/", URLShortener(service))
	r.Get("/{id}", URLRedirect(service))

	return r
}
