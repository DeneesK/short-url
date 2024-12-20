package main

import (
	"log"
	"net/http"

	"github.com/DeneesK/short-url/internal/app/repository"
	"github.com/DeneesK/short-url/internal/app/router"
)

func main() {
	rep := repository.NewRepository()
	router := router.NewRouter(rep)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
