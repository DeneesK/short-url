package middlewares

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const (
	cookieName              = "user"
	cookieMaxAge            = 3600 * 24 * 30 * time.Second
	UserIDKey    contextKey = "userID"
)

type UserService interface {
	Create(ctx context.Context) (string, error)
	Verify(user string) bool
}

func NewUserCookieMiddleware(log Logger, userService UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			_, err := r.Cookie(cookieName)
			if err == http.ErrNoCookie {
				user, err := userService.Create(r.Context())
				if err != nil {
					log.Errorf("failed to create user %s", err)
					http.Error(w, "failed to create user", http.StatusBadRequest)
					return
				}
				setCookie(w, user)
				values := strings.Split(user, ":")
				userID := values[0]
				ctx = context.WithValue(r.Context(), UserIDKey, userID)
			} else if err != nil {
				log.Errorf("failed to get cookie %s", err)
				http.Error(w, "failed request", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func NewUserVerifyMiddleware(log Logger, userService UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := r.Cookie(cookieName)
			if err != nil {
				log.Errorf("failed to get cookie %s", err)
				http.Error(w, "failed request", http.StatusBadRequest)
				return
			}
			if !userService.Verify(user.Value) {
				log.Errorf("failed to verify user %s", user.Value)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			setCookie(w, user.Value) // обновляем ttl

			values := strings.Split(user.Value, ":")
			userID := values[0]
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func setCookie(w http.ResponseWriter, value string) {
	cookie := &http.Cookie{
		Name:    cookieName,
		Value:   value,
		Expires: time.Now().Add(cookieMaxAge),
	}
	http.SetCookie(w, cookie)
}
