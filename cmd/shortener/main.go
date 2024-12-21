package main

import (
	"log"
	"net/http"

	"github.com/DeneesK/short-url/internal/app/conf"
	"github.com/DeneesK/short-url/internal/app/repository"
	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/DeneesK/short-url/internal/storage"
)

func main() {
	conf := conf.MustLoad()
	storage := storage.NewMemoryStorage()
	rep := repository.NewRepository(storage, conf.BaseURL)
	router := router.NewRouter(rep)
	if err := http.ListenAndServe(conf.ServerAddr, router); err != nil {
		log.Fatal(err)
	}
}
