package main

import (
	"github.com/DeneesK/short-url/internal/app"
	"github.com/DeneesK/short-url/internal/app/conf"
	"github.com/DeneesK/short-url/internal/app/logger"
	"github.com/DeneesK/short-url/internal/app/repository"
	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/DeneesK/short-url/internal/app/service"
	"github.com/DeneesK/short-url/internal/app/storage/memorystorage"
)

func main() {
	conf := conf.MustLoad()

	log := logger.NewLogger(conf.Env)
	defer log.Sync()
	storage := memorystorage.NewMemoryStorage(conf.MemoryUsageLimitBytes)
	rep, err := repository.NewRepository(
		storage,
		conf.DBDSN,
		repository.AddDumpFile(conf.FileStoragePath),
		repository.RestoreFromDump(conf.FileStoragePath),
	)
	if err != nil {
		log.Fatalf("failed to initialized repository: %s", err)
	}
	defer rep.Close()

	service := service.NewURLShortener(rep, conf.BaseURL)
	router := router.NewRouter(service, log)

	app := app.NewApp(conf.ServerAddr, router)
	app.Run()
}
