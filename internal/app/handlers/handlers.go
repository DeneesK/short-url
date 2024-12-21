package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const baseRespURL = "http://localhost:8080/"

type URLRepository interface {
	SaveURL(string) (string, error)
	GetURL(string) (string, error)
}

func URLSaver(urlSaver URLRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request's body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if string(body) == "" {
			http.Error(w, "body must have url", http.StatusBadRequest)
			return
		}

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

func URLRedirect(urlStorage URLRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if id == "" {
			http.Error(w, "ID not provided", http.StatusBadRequest)
			return
		}

		url, err := urlStorage.GetURL(id)
		if err != nil {
			errorString := fmt.Sprintf("failed to redirect: %s", err.Error())
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
