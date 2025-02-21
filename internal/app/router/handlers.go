package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/services"
	"github.com/go-chi/chi/v5"
)

const (
	cookieName = "user"
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

		LongURL := string(body)
		user, err := r.Cookie(cookieName)
		if err != nil {
			log.Errorf("failed request %s", err)
			http.Error(w, "failed request", http.StatusBadRequest)
			return
		}
		values := strings.Split(user.Value, ":")
		userID := values[0]
		shortURL, err := urlService.ShortenURL(r.Context(), LongURL, userID)
		if err != nil && err != services.ErrLongURLAlreadyExists {
			errorString := fmt.Sprintf("failed to create short url: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		} else if err == services.ErrLongURLAlreadyExists {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
		}
		w.Write([]byte(shortURL))
	}
}

func URLShortenerJSON(urlService URLService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var LongURL LongURL

		err := json.NewDecoder(r.Body).Decode(&LongURL)
		if err != nil {
			log.Errorf("failed to decode request's body %s", err)
			http.Error(w, "failed to decode request's body", http.StatusBadRequest)
			return
		}
		user, err := r.Cookie(cookieName)
		if err != nil {
			log.Errorf("failed request %s", err)
			http.Error(w, "failed request", http.StatusBadRequest)
			return
		}
		values := strings.Split(user.Value, ":")
		userID := values[0]
		shortURL, err := urlService.ShortenURL(r.Context(), LongURL.URL, userID)
		if err != nil && err != services.ErrLongURLAlreadyExists {
			errorString := fmt.Sprintf("failed to create short url: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		} else if err == services.ErrLongURLAlreadyExists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
		}
		res := ShortURL{Result: shortURL}

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
		if url.IsDeleted {
			w.WriteHeader(http.StatusGone)
			return
		}
		if err != nil {
			errorString := fmt.Sprintf("failed to redirect: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url.LongURL)
		http.Redirect(w, r, url.LongURL, http.StatusTemporaryRedirect)
	}
}

func URLShortenerBatchJSON(urlService URLService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		batch := make([]dto.OriginalURL, 0)

		err := json.NewDecoder(r.Body).Decode(&batch)
		if err != nil {
			log.Errorf("failed to decode request's body %s", err)
			http.Error(w, "failed to decode request's body", http.StatusBadRequest)
			return
		}
		user, err := r.Cookie(cookieName)
		if err != nil {
			log.Errorf("failed request %s", err)
			http.Error(w, "failed request", http.StatusBadRequest)
			return
		}
		values := strings.Split(user.Value, ":")
		userID := values[0]
		result, err := urlService.StoreBatchURL(r.Context(), batch, userID)
		if err != nil {
			errorString := fmt.Sprintf("failed to create short url: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			errorString := fmt.Sprintf("failed to encode short url: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}
	}
}

func URLsByUser(urlService URLService, userService UserService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := r.Cookie(cookieName)
		if err != nil {
			log.Errorf("failed to get cookie %s", err)
			http.Error(w, "failed request", http.StatusBadRequest)
			return
		}
		values := strings.Split(user.Value, ":")
		userID := values[0]
		urls, err := urlService.FindByUserID(r.Context(), userID)
		if err != nil {
			log.Errorf("failed request %s", err)
			http.Error(w, "failed request", http.StatusBadRequest)
			return
		}
		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(urls)
		if err != nil {
			errorString := fmt.Sprintf("failed to encode: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
		}
	}
}

func DeleteByAlias(urlService URLService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var idx []string

		err := json.NewDecoder(r.Body).Decode(&idx)
		if err != nil {
			log.Errorf("failed to decode request's body %s", err)
			http.Error(w, "failed to decode request's body", http.StatusBadRequest)
			return
		}
		user, err := r.Cookie(cookieName)
		if err != nil {
			log.Errorf("failed request %s", err)
			http.Error(w, "failed request", http.StatusBadRequest)
			return
		}

		values := strings.Split(user.Value, ":")
		userID := values[0]
		urlService.DeleteBatch(idx, userID)
		w.WriteHeader(http.StatusAccepted)
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
