package middlewares

import (
	"context"
	"net/http"
	"time"
)

const (
	cookieName   = "user"
	cookieMaxAge = 3600 * 24 * 30
)

type UserService interface {
	Create(ctx context.Context) (string, error)
	Verify(user string) bool
}

func NewUserCookieMiddleware(log Logger, userService UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := r.Cookie(cookieName)
			if err == http.ErrNoCookie {
				user, err := userService.Create(r.Context())
				if err != nil {
					log.Errorf("failed to create user %s", err)
					http.Error(w, "ailed to create user", http.StatusBadRequest)
					return
				}
				setCookie(w, r, user)
			} else if err != nil {
				log.Errorf("failed to get cookie %s", err)
				http.Error(w, "failed request", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
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
			next.ServeHTTP(w, r)
		})
	}
}

func setCookie(w http.ResponseWriter, r *http.Request, value string) {
	cookie := &http.Cookie{
		Name:    cookieName,
		Value:   value,
		Expires: time.Now().Add(time.Duration(cookieMaxAge) * time.Second),
	}
	http.SetCookie(w, cookie)
	r.AddCookie(cookie)
}
