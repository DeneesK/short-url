package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/storage"
	"github.com/DeneesK/short-url/pkg/random"
	"github.com/DeneesK/short-url/pkg/validator"
)

const (
	maxRetries = 3
	idLength   = 8
)

type Repository interface {
	Store(ctx context.Context, id, value string) error
	StoreBatch(context.Context, [][2]string) error
	Get(ctx context.Context, id string) (string, error)
	PingDB(ctx context.Context) error
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

func (s *URLShortener) ShortenURL(ctx context.Context, longURL string) (string, error) {
	var alias string
	var err error
	if isValid := validator.IsValidURL(longURL); !isValid {
		return "", fmt.Errorf("this url: '%s' is not valid url", longURL)
	}

	for i := 0; i < maxRetries; i++ {
		alias = random.RandomString(idLength)

		err = s.rep.Store(ctx, alias, longURL)
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

func (s *URLShortener) FindByShortened(ctx context.Context, id string) (string, error) {
	shortURL, err := s.rep.Get(ctx, id)
	if err != nil {
		return "", nil
	}
	return shortURL, nil
}

func (s *URLShortener) StoreBatchURL(ctx context.Context, batch []dto.OriginalURL) ([]dto.ShortedURL, error) {
	result := make([]dto.ShortedURL, 0, len(batch))
	chunk := make([][2]string, 0, len(batch))
	for _, origin := range batch {
		if isValid := validator.IsValidURL(origin.URL); !isValid {
			return nil, fmt.Errorf("this url: '%s' is not valid url", origin.URL)
		}
		chunk = append(chunk, [2]string{origin.ID, origin.URL})
		shortURL, err := url.JoinPath(s.baseAddr, origin.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, dto.ShortedURL{ID: origin.ID, URL: shortURL})
	}

	err := s.rep.StoreBatch(ctx, chunk)
	if err != nil {
		return nil, err
	}

	return result, nil
}
func (s *URLShortener) PingDB(ctx context.Context) error {
	return s.rep.PingDB(ctx)
}
