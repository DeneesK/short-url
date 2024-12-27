package repository

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/DeneesK/short-url/internal/pkg/random"
	"github.com/DeneesK/short-url/internal/storage"
)

const (
	maxRetries = 3
	idLength   = 8
)

type Storage interface {
	Save(id, value string) error
	Get(id string) (string, error)
}

type Repository struct {
	storage  Storage
	baseAddr string
}

func (r *Repository) SaveURL(u string) (string, error) {
	var alias string
	var err error

	for i := 0; i < maxRetries; i++ {
		alias = random.RandomString(idLength)

		err = r.storage.Save(alias, u)
		if err != nil {
			if errors.Is(err, storage.ErrNotUniqueID) {
				continue
			}
			return "", fmt.Errorf("failed to generate unique alias after %d attempts", maxRetries)
		}
		break
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

func NewRepository(storage Storage, baseAddr string) *Repository {
	return &Repository{
		storage:  storage,
		baseAddr: baseAddr,
	}
}
