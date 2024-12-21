package main

import (
	"log"
	"net/http"

	"github.com/DeneesK/short-url/internal/app/repository"
	"github.com/DeneesK/short-url/internal/app/router"
	"github.com/DeneesK/short-url/internal/storage"
)

func main() {
	storage := storage.NewMemoryStorage()
	rep := repository.NewRepository(storage)
	router := router.NewRouter(rep)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
