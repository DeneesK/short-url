package main

import (
	"context"

	"github.com/DeneesK/short-url/internal/app"
	"github.com/DeneesK/short-url/internal/app/conf"
	"github.com/DeneesK/short-url/internal/app/logger"
	"github.com/DeneesK/short-url/internal/app/repository"
	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/DeneesK/short-url/internal/app/services"
)

func main() {
	conf := conf.MustLoad()

	log := logger.NewLogger(conf.Env)
	defer log.Sync()
	rep, err := repository.NewRepository(
		repository.StorageConfig{
			DBDSN:           conf.DBDSN,
			MaxStorageSize:  conf.MemoryUsageLimitBytes,
			MigrationSource: conf.MigrationsPath,
		},
		repository.AddDumpFile(conf.FileStoragePath),
		repository.RestoreFromDump(conf.FileStoragePath),
	)
	if err != nil {
		log.Fatalf("failed to initialized repository: %s", err)
	}
	ctx, close := context.WithCancel(context.Background())
	defer close()
	defer rep.Close(ctx)

	urlService := services.NewURLShortener(rep, conf.BaseURL)
	userService := services.NewUserService(rep, conf.SecretKey)
	router := router.NewRouter(urlService, userService, log)

	app := app.NewApp(conf.ServerAddr, router, log)
	app.Run()
}
