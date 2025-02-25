package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/router"
)

const shutdownTimeout = time.Second * 1

type Logger interface {
	Infoln(args ...interface{})
	Fatalf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Error(args ...interface{})
}

type URLService interface {
	ShortenURL(context.Context, string, string) (string, error)
	StoreBatchURL(context.Context, []dto.OriginalURL, string) ([]dto.ShortedURL, error)
	FindByShortened(context.Context, string) (dto.LongURL, error)
	FindByUserID(context.Context, string) ([]dto.URL, error)
	DeleteBatch([]string, string)
	PingDB(context.Context) error
	Shutdown()
}

type UserService interface {
	Create(ctx context.Context) (string, error)
	Verify(user string) bool
}

type APP struct {
	srv         *http.Server
	urlService  URLService
	userService UserService
	log         Logger
}

func NewApp(addr string, urlService URLService, userService UserService, log Logger) *APP {
	r := router.NewRouter(urlService, userService, log)
	s := http.Server{
		Addr:    addr,
		Handler: r,
	}
	return &APP{
		srv:         &s,
		log:         log,
		urlService:  urlService,
		userService: userService,
	}
}

func (a *APP) Run() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)
	defer stop()

	a.log.Infoln("starting app, server listening on", a.srv.Addr)

	go func() {
		err := a.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			a.log.Fatalf("failed to start server: %s", err)
		}
	}()

	<-ctx.Done()

	a.urlService.Shutdown()
	a.log.Infoln("application shutdown process...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error during shutdown: %s", err)
	}
	<-shutdownCtx.Done()
	a.log.Infoln("application and server gracefully stopped")
}
