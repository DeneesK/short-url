package main

import (
	"net/http"

	"github.com/DeneesK/short-url/internal/app/conf"
	"github.com/DeneesK/short-url/internal/app/logger"
	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/DeneesK/short-url/internal/app/service"
	"github.com/DeneesK/short-url/internal/app/storage/memorystorage"
)

func main() {
	conf := conf.MustLoad()
	log := logger.MustInitializedLogger(conf.Env)
	defer log.Sync()

	storage := memorystorage.NewMemoryStorage(conf.MemoryUsageLimitBytes)
	service := service.NewURLShortener(storage, conf.BaseURL)
	router := router.NewRouter(service, log)

	log.Infof("starting server, listening %s", conf.ServerAddr)

	if err := http.ListenAndServe(conf.ServerAddr, router); err != nil {
		log.Fatalw(err.Error(), "event", "start server")
	}
}
