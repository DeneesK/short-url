package service

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/DeneesK/short-url/internal/app/storage"
	"github.com/DeneesK/short-url/pkg/random"
	"github.com/DeneesK/short-url/pkg/validator"
)

const (
	maxRetries = 3
	idLength   = 8
)

type Repository interface {
	Store(id, value string) error
	Get(id string) (string, error)
	PingDB() error
}

type URLShortener struct {
	rep      Repository
	baseAddr string
}

func NewURLShortener(storage Repository, baseAddr string) *URLShortener {
	return &URLShortener{
		rep:      storage,
		baseAddr: baseAddr,
	}
}

func (s *URLShortener) ShortenURL(longURL string) (string, error) {
	var alias string
	var err error
	if isValid := validator.IsValidURL(longURL); !isValid {
		return "", fmt.Errorf("this url: '%s' is not valid url", longURL)
	}

	for i := 0; i < maxRetries; i++ {
		alias = random.RandomString(idLength)

		err = s.rep.Store(alias, longURL)
		if err != nil {
			if errors.Is(err, storage.ErrNotUniqueID) {
				continue
			} else {
				return "", err
			}
		}
		break
	}
	if err != nil {
		return "", fmt.Errorf("failed to generate unique alias after %d attempts", maxRetries)
	}

	shortURL, err := url.JoinPath(s.baseAddr, alias)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (s *URLShortener) FindByShortened(id string) (string, error) {
	shortURL, err := s.rep.Get(id)
	if err != nil {
		return "", nil
	}
	return shortURL, nil
}

func (s *URLShortener) PingDB() error {
	return s.rep.PingDB()
}
