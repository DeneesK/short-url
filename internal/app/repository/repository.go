package repository

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/DeneesK/short-url/internal/pkg/random"
	"github.com/DeneesK/short-url/internal/storage"
)

var ErrStorageLimitExceeded = errors.New("storage limit exceeded")

const (
	maxRetries = 3
	idLength   = 8
)

type Storage interface {
	Save(id, value string) error
	Get(id string) (string, error)
	CurrentSize() uint64
}

type Repository struct {
	storage        Storage
	baseAddr       string
	maxStorageSize uint64
}

func (r *Repository) SaveURL(u string) (string, error) {
	if r.storage.CurrentSize() > r.maxStorageSize {
		return "", ErrStorageLimitExceeded
	}

	var alias string
	var err error

	for i := 0; i < maxRetries; i++ {
		alias = random.RandomString(idLength)

		err = r.storage.Save(alias, u)
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

	shortURL, err := url.JoinPath(r.baseAddr, alias)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (r *Repository) GetURL(id string) (string, error) {
	url, err := r.storage.Get(id)
	if err != nil {
		return "", err
	}
	return url, nil
}

func NewRepository(storage Storage, baseAddr string, maxStorageSize uint64) *Repository {
	return &Repository{
		storage:        storage,
		baseAddr:       baseAddr,
		maxStorageSize: maxStorageSize,
	}
}
