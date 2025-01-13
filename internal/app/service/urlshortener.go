package service

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/DeneesK/short-url/internal/app/storage"
	"github.com/DeneesK/short-url/pkg/random"
)

const (
	maxRetries = 3
	idLength   = 8
)

type Storage interface {
	Store(id, value string) error
	Get(id string) (string, error)
}

type URLShortener struct {
	storage  Storage
	baseAddr string
}

func (s *URLShortener) ShortenURL(longURL string) (string, error) {
	var alias string
	var err error

	for i := 0; i < maxRetries; i++ {
		alias = random.RandomString(idLength)

		err = s.storage.Store(alias, longURL)
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
	shortURL, err := s.storage.Get(id)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func NewURLShortener(storage Storage, baseAddr string) *URLShortener {
	return &URLShortener{
		storage:  storage,
		baseAddr: baseAddr,
	}
}
