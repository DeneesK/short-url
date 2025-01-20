package router

import (
	"github.com/go-chi/chi/v5"
)

type URLService interface {
	ShortenURL(string) (string, error)
	FindByShortened(string) (string, error)
}

type Logger interface {
	Infoln(args ...interface{})
	Errorf(template string, args ...interface{})
	Error(args ...interface{})
}

func NewRouter(service URLService, log Logger) *chi.Mux {
	r := chi.NewRouter()

	loggingMiddleware := NewLoggingMiddleware(log)
	gzipMiddleware := NewGZIPMiddleware(log)
	r.Use(loggingMiddleware, gzipMiddleware)

	r.Post("/", URLShortener(service, log))
	r.Post("/api/shorten", URLShortenerJSON(service, log))
	r.Get("/{id}", URLRedirect(service, log))

	return r
}
