package router

import (
	"context"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/router/middlewares"
	"github.com/go-chi/chi/v5"
)

type URLService interface {
	ShortenURL(context.Context, string) (string, error)
	StoreBatchURL(context.Context, []dto.OriginalURL) ([]dto.ShortedURL, error)
	FindByShortened(context.Context, string) (string, error)
	PingDB(context.Context) error
}

type Logger interface {
	Infoln(args ...interface{})
	Errorf(template string, args ...interface{})
	Error(args ...interface{})
}

func NewRouter(service URLService, log Logger) *chi.Mux {
	r := chi.NewRouter()

	loggingMiddleware := middlewares.NewLoggingMiddleware(log)
	gzipReqDecodeMiddleware := middlewares.NewRequestDecodeMiddleware(log)
	gzipRespEncodeMiddleware := middlewares.NewResponseEncodeMiddleware(log)
	r.Use(loggingMiddleware, gzipReqDecodeMiddleware, gzipRespEncodeMiddleware)

	r.Post("/", URLShortener(service, log))
	r.Post("/api/shorten/batch", URLShortenerBatchJSON(service, log))
	r.Post("/api/shorten", URLShortenerJSON(service, log))
	r.Get("/{id}", URLRedirect(service, log))
	r.Get("/ping", PingDB(service, log))

	return r
}
