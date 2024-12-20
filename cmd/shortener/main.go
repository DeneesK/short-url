package main

import (
	"log"
	"net/http"

	"github.com/DeneesK/short-url/internal/app/repository"
	"github.com/DeneesK/short-url/internal/app/router"
)

func main() {
	s := repository.NewRepository()
	r := router.NewRouter(s)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
