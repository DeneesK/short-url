package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 1

type Logger interface {
	Fatalf(template string, args ...interface{})
	Infoln(args ...interface{})
}

type APP struct {
	srv *http.Server
	log Logger
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
			log.Fatalf("failed to start server: %s", err)
		}
	}()

	<-ctx.Done()
	a.log.Infoln("application shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*shutdownTimeout)
	defer cancel()
	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error during shutdown: %s", err)
	}
	<-shutdownCtx.Done()
	a.log.Infoln("application and server gracefully stopped")
}

func NewApp(addr string, handler http.Handler, log Logger) *APP {
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	return &APP{srv: &s, log: log}
}
