package repository

import (
	"net/url"

	"github.com/DeneesK/short-url/internal/pkg/random"
)

const idLength = 8

type Storage interface {
	Save(id, value string) (string, error)
	Get(id string) (string, error)
}

type Repository struct {
	storage  Storage
	baseAddr string
}

func (r *Repository) SaveURL(u string) (string, error) {
	id := random.RandomString(idLength)
	r.storage.Save(id, u)
	shortURL, err := url.JoinPath(r.baseAddr, id)
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
