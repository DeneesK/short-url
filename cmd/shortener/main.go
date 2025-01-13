package main

import (
	"log"
	"net/http"

	"github.com/DeneesK/short-url/internal/app/conf"
	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/DeneesK/short-url/internal/app/service"
	memstorage "github.com/DeneesK/short-url/internal/app/storage/memory_storage"
)

func main() {
	conf := conf.MustLoad()
	storage := memstorage.NewMemoryStorage(conf.MemoryUsageLimitBytes)
	service := service.NewURLShortener(storage, conf.BaseURL)
	router := router.NewRouter(service)

	if err := http.ListenAndServe(conf.ServerAddr, router); err != nil {
		log.Fatal(err)
	}
}
