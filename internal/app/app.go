package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = time.Second * 1

type APP struct {
	srv *http.Server
}

func NewApp(addr string, handler http.Handler) *APP {
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	return &APP{srv: &s}
}

func (a *APP) Run() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)
	defer stop()

	log.Println("starting app, server listening on", a.srv.Addr)

	go func() {
		err := a.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %s", err)
		}
	}()

	<-ctx.Done()

	log.Print("application shutdown process...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error during shutdown: %s", err)
	}
	<-shutdownCtx.Done()
	log.Print("application and server gracefully stopped")
}
