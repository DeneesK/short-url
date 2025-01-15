package router

import (
	"fmt"
	"io"
	"net/http"

	"github.com/DeneesK/short-url/pkg/validator"
	"github.com/go-chi/chi/v5"
)

func URLShortener(urlService URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request's body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		url := string(body)

		if isValid := validator.IsValidURL(url); !isValid {
			http.Error(w, "body must have valid url", http.StatusBadRequest)
			return
		}

		shortURL, err := urlService.ShortenURL(url)
		if err != nil {
			errorString := fmt.Sprintf("failed to create short url: %s", err.Error())
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func URLRedirect(urlStorage URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if id == "" {
			http.Error(w, "ID not provided", http.StatusBadRequest)
			return
		}

		url, err := urlStorage.FindByShortened(id)
		if err != nil {
			errorString := fmt.Sprintf("failed to redirect: %s", err.Error())
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
