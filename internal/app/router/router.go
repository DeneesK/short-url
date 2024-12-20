package router

import (
	"net/http"

	"github.com/DeneesK/short-url/internal/app/handlers"
)

type Repository interface {
	SaveURL(string) (string, error)
	GetURL(string) (string, error)
}

func NewRouter(rep Repository) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.URLHandler(rep))
	return mux
}
