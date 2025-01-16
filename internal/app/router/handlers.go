package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DeneesK/short-url/pkg/validator"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type LongURL struct {
	Url string `json:"url"`
}

type ShortURL struct {
	Result string `json:"result"`
}

func URLShortener(urlService URLService, log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf("failed to read request's body %s", err)
			http.Error(w, "failed to read request's body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		url := string(body)

		if isValid := validator.IsValidURL(url); !isValid {
			log.Errorf("body must have valid url, url %s", url)
			http.Error(w, "body must have valid url", http.StatusBadRequest)
			return
		}

		shortURL, err := urlService.ShortenURL(url)
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

func URLShortenerJSON(urlService URLService, log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var longURL LongURL

		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&longURL)
		if err != nil {
			log.Errorf("failed to decode request's body %s", err)
			http.Error(w, "failed to decode request's body", http.StatusBadRequest)
			return
		}

		if isValid := validator.IsValidURL(longURL.Url); !isValid {
			log.Errorf("body must have valid url, url %s", longURL.Url)
			http.Error(w, "body must have valid url", http.StatusBadRequest)
			return
		}

		shortURL, err := urlService.ShortenURL(longURL.Url)
		if err != nil {
			errorString := fmt.Sprintf("failed to create short url: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		res := ShortURL{Result: shortURL}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		err = enc.Encode(res)
		if err != nil {
			errorString := fmt.Sprintf("failed to encode short url: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}
	}
}

func URLRedirect(urlService URLService, log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if id == "" {
			log.Error("ID not provided")
			http.Error(w, "ID not provided", http.StatusBadRequest)
			return
		}

		url, err := urlService.FindByShortened(id)
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
