package main

import (
	"net/http"

	"github.com/DeneesK/short-url/internal/app/conf"
	"github.com/DeneesK/short-url/internal/app/logger"
	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/DeneesK/short-url/internal/app/service"
	memstorage "github.com/DeneesK/short-url/internal/app/storage/memory_storage"
)

func main() {
	conf := conf.MustLoad()
	log := logger.MustInitializedLogger()
	storage := memstorage.NewMemoryStorage(conf.MemoryUsageLimitBytes)
	service := service.NewURLShortener(storage, conf.BaseURL)
	router := router.NewRouter(service, log)

	log.Infof("starting server, listening %s", conf.ServerAddr)

	if err := http.ListenAndServe(conf.ServerAddr, router); err != nil {
		log.Fatalw(err.Error(), "event", "start server")
	}
}
