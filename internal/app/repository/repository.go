package repository

import (
	"fmt"

	"github.com/DeneesK/short-url/internal/pkg/random"
)

const idLength = 8

type Storage interface {
	Save(id, value string) (string, error)
	Get(id string) (string, error)
}

type Repository struct {
	storage Storage
}

func (r *Repository) SaveURL(url string) (string, error) {
	id := random.RandomString(idLength)
	r.storage.Save(id, url)
	return id, nil
}

func (r *Repository) GetURL(id string) (string, error) {
	url, err := r.storage.Get(id)
	if err != nil {
		return "", fmt.Errorf("url not found by id: %v", id)
	}
	return url, nil
}

func NewRepository(storage Storage) *Repository {
	return &Repository{
		storage: storage,
	}
}
