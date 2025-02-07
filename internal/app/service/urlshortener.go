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

var ErrLongURLAlreadyExists = errors.New("long URL already exists")

const (
	maxRetries = 3
	idLength   = 8
)

type Repository interface {
	Store(context.Context, string, string) (string, error)
	StoreBatch(context.Context, [][2]string) error
	Get(context.Context, string) (string, error)
	PingDB(context.Context) error
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

		alias, err = s.rep.Store(ctx, alias, longURL)
		if err != nil {
			if errors.Is(err, storage.ErrNotUniqueID) {
				continue
			} else if errors.Is(err, storage.ErrUniqueViolation) {
				shortURL, err := url.JoinPath(s.baseAddr, alias)
				if err != nil {
					return "", err
				}
				return shortURL, ErrLongURLAlreadyExists
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
	data := make([][2]string, 0, len(batch))
	for _, origin := range batch {
		if isValid := validator.IsValidURL(origin.URL); !isValid {
			return nil, fmt.Errorf("this url: '%s' is not valid url", origin.URL)
		}
		data = append(data, [2]string{origin.ID, origin.URL})
		shortURL, err := url.JoinPath(s.baseAddr, origin.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, dto.ShortedURL{ID: origin.ID, URL: shortURL})
	}

	err := s.rep.StoreBatch(ctx, data)
	if err != nil {
		return nil, err
	}

	return result, nil
}
func (s *URLShortener) PingDB(ctx context.Context) error {
	return s.rep.PingDB(ctx)
}
