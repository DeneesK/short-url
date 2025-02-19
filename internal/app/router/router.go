package router

import (
	"context"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/router/middlewares"
	"github.com/go-chi/chi/v5"
)

type URLService interface {
	ShortenURL(context.Context, string, string) (string, error)
	StoreBatchURL(context.Context, []dto.OriginalURL, string) ([]dto.ShortedURL, error)
	FindByShortened(context.Context, string) (string, error)
	FindByUserID(context.Context, string) ([]dto.URL, error)
	PingDB(context.Context) error
}

type UserService interface {
	Create(ctx context.Context) (string, error)
	Verify(user string) bool
}

type Logger interface {
	Infoln(args ...interface{})
	Errorf(template string, args ...interface{})
	Error(args ...interface{})
}

func NewRouter(urlService URLService, userService UserService, log Logger) *chi.Mux {
	r := chi.NewRouter()

	loggingMiddleware := middlewares.NewLoggingMiddleware(log)
	gzipReqDecodeMiddleware := middlewares.NewRequestDecodeMiddleware(log)
	gzipRespEncodeMiddleware := middlewares.NewResponseEncodeMiddleware(log)
	userCookieMiddleware := middlewares.NewUserCookieMiddleware(log, userService)
	userVerifier := middlewares.NewUserVerifyMiddleware(log, userService)
	r.Use(loggingMiddleware, gzipReqDecodeMiddleware, gzipRespEncodeMiddleware, userCookieMiddleware, userVerifier)

	r.Post("/", URLShortener(urlService, log))
	r.Post("/api/shorten/batch", URLShortenerBatchJSON(urlService, log))
	r.Post("/api/shorten", URLShortenerJSON(urlService, log))

	r.Get("/{id}", URLRedirect(urlService, log))
	r.Get("/ping", PingDB(urlService, log))
	r.Get("/api/user/urls", URLsByUser(urlService, userService, log))

	return r
}
