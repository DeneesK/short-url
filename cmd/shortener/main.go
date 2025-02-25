package main

import (
	"context"

	"github.com/DeneesK/short-url/internal/app"
	"github.com/DeneesK/short-url/internal/app/conf"
	"github.com/DeneesK/short-url/internal/app/logger"
	"github.com/DeneesK/short-url/internal/app/repository"
	"github.com/DeneesK/short-url/internal/app/service"
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

	urlService := service.NewURLShortener(rep, conf.BaseURL, log)
	userService := service.NewUserService(rep, conf.SecretKey)

	app := app.NewApp(conf.ServerAddr, urlService, userService, log)
	app.Run()
}
