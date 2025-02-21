package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/storage"
	"github.com/DeneesK/short-url/pkg/random"
	"github.com/DeneesK/short-url/pkg/validator"
	"go.uber.org/zap"
)

var ErrLongURLAlreadyExists = errors.New("long URL already exists")

const (
	maxRetries = 3
	numWorkers = 3
	idLength   = 8
	sleepTime  = 100
)

type Repository interface {
	Store(ctx context.Context, id, value, userID string) (string, error)
	StoreBatch(ctx context.Context, batch []dto.OriginalURL, userID string) error
	UpdateStatusBatch([]dto.UpdateTask) error
	GetByUserID(ctx context.Context, userID string) ([]dto.OriginalURL, error)
	Get(context.Context, string) (dto.LongURL, error)
	PingDB(context.Context) error
}

type URLShortener struct {
	rep      Repository
	baseAddr string
	log      *zap.SugaredLogger
}

func NewURLShortener(storage Repository, baseAddr string, log *zap.SugaredLogger) *URLShortener {
	return &URLShortener{
		rep:      storage,
		baseAddr: baseAddr,
		log:      log,
	}
}

func (s *URLShortener) ShortenURL(ctx context.Context, LongURL, userID string) (string, error) {
	var alias string
	var err error
	if isValid := validator.IsValidURL(LongURL); !isValid {
		return "", fmt.Errorf("this url: '%s' is not valid url", LongURL)
	}

	for i := 0; i < maxRetries; i++ {
		alias = random.RandomString(idLength)

		alias, err = s.rep.Store(ctx, alias, LongURL, userID)
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
				time.Sleep(sleepTime * time.Millisecond)
				continue
			}
		}
		break
	}
	if err != nil {
		return "", fmt.Errorf("failed to store shorten URL after %d attempts, reason: %q", maxRetries, err)
	}

	shortURL, err := url.JoinPath(s.baseAddr, alias)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (s *URLShortener) FindByShortened(ctx context.Context, id string) (dto.LongURL, error) {
	originalURL, err := s.rep.Get(ctx, id)
	if err != nil {
		return dto.LongURL{}, err
	}
	return originalURL, nil
}

func (s *URLShortener) FindByUserID(ctx context.Context, userID string) ([]dto.URL, error) {
	originalURLs, err := s.rep.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	urls := make([]dto.URL, 0)
	for _, original := range originalURLs {
		shortURL, err := url.JoinPath(s.baseAddr, original.ID)
		if err != nil {
			return nil, err
		}
		u := dto.URL{
			OriginalURL: original.URL,
			ShortURL:    shortURL,
		}
		urls = append(urls, u)
	}
	return urls, nil
}

func (s *URLShortener) StoreBatchURL(ctx context.Context, batch []dto.OriginalURL, userID string) ([]dto.ShortedURL, error) {
	result := make([]dto.ShortedURL, 0, len(batch))
	for _, origin := range batch {
		if isValid := validator.IsValidURL(origin.URL); !isValid {
			return nil, fmt.Errorf("%q is not valid url", origin.URL)
		}
		shortURL, err := url.JoinPath(s.baseAddr, origin.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, dto.ShortedURL{ID: origin.ID, URL: shortURL})
	}

	err := s.rep.StoreBatch(ctx, batch, userID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *URLShortener) DeleteBatch(idx []string, userID string) {
	updateCh := s.addToUpdate(idx, userID)
	s.startStatusUpdater(updateCh)

}

func (s *URLShortener) PingDB(ctx context.Context) error {
	return s.rep.PingDB(ctx)
}

func (s *URLShortener) startStatusUpdater(updateCh chan dto.UpdateTask) {
	for i := 0; i < numWorkers; i++ {
		go s.executeBatchUpdate(updateCh)
	}
}

func (s *URLShortener) addToUpdate(idx []string, userID string) chan dto.UpdateTask {
	updateCh := make(chan dto.UpdateTask, len(idx))
	defer close(updateCh)
	for _, id := range idx {
		updateCh <- dto.UpdateTask{UserID: userID, ID: id}
	}
	return updateCh
}

func (s *URLShortener) executeBatchUpdate(updateCh chan dto.UpdateTask) {
	batch := make([]dto.UpdateTask, 0)
	for task := range updateCh {
		batch = append(batch, task)
	}
	err := s.rep.UpdateStatusBatch(batch)
	if err != nil {
		s.log.Errorf("during attempt of batch status updating: %s", err.Error())
	}
}
