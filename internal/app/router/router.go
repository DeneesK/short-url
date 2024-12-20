package router

import (
	"net/http"

	"github.com/DeneesK/short-url/internal/app/handlers"
)

type Repository interface {
	SaveUrl(string) (string, error)
	GetUrl(string) (string, error)
}

func NewRouter(rep Repository) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.UrlHandler(rep))
	return mux
}
