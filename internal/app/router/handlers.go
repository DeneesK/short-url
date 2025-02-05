package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type LongURL struct {
	URL string `json:"url"`
}

type ShortURL struct {
	Result string `json:"result"`
}

func URLShortener(urlService URLService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf("failed to read request's body %s", err)
			http.Error(w, "failed to read request's body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		longURL := string(body)

		shortURL, err := urlService.ShortenURL(r.Context(), longURL)
		if err != nil {
			errorString := fmt.Sprintf("failed to create short url: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func URLShortenerJSON(urlService URLService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var longURL LongURL

		err := json.NewDecoder(r.Body).Decode(&longURL)
		if err != nil {
			log.Errorf("failed to decode request's body %s", err)
			http.Error(w, "failed to decode request's body", http.StatusBadRequest)
			return
		}

		shortURL, err := urlService.ShortenURL(r.Context(), longURL.URL)
		if err != nil {
			errorString := fmt.Sprintf("failed to create short url: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		res := ShortURL{Result: shortURL}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			errorString := fmt.Sprintf("failed to encode short url: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}
	}
}

func URLRedirect(urlService URLService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if id == "" {
			log.Error("ID not provided")
			http.Error(w, "ID not provided", http.StatusBadRequest)
			return
		}

		url, err := urlService.FindByShortened(r.Context(), id)
		if err != nil {
			errorString := fmt.Sprintf("failed to redirect: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func PingDB(urlService URLService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := urlService.PingDB(r.Context())
		if err != nil {
			log.Errorf("failed to ping db: %s", err)
			http.Error(w, "database is not available", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
