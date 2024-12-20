package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseRespURL = "http://localhost:8080/"

type URLRepository interface {
	SaveURL(string) (string, error)
	GetURL(string) (string, error)
}

func URLHandler(urlSaver URLRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			parts := strings.Split(r.URL.Path, "/")

			if len(parts) < 2 || parts[1] == "" {
				http.Error(w, "ID not provided", http.StatusBadRequest)
				return
			}

			id := parts[1]

			url, err := urlSaver.GetURL(id)
			if err != nil {
				errorString := fmt.Sprintf("failed to redirect: %s", err.Error())
				http.Error(w, errorString, http.StatusBadRequest)
				return
			}
			w.Header().Set("Location", url)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request's body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		id, err := urlSaver.SaveURL(string(body))
		if err != nil {
			errorString := fmt.Sprintf("failed to create short url: %s", err.Error())
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		shortURL := baseRespURL + id

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}
