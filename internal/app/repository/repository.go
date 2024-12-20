package repository

import (
	"fmt"
	"sync"

	"github.com/DeneesK/short-url/internal/pkg/random"
)

const idLength = 8

type Repository struct {
	m       sync.RWMutex
	storage map[string]string
}

func (r *Repository) SaveURL(url string) (string, error) {
	r.m.Lock()
	defer r.m.Unlock()
	id := random.RandomString(idLength)
	r.storage[id] = url
	return id, nil
}

func (r *Repository) GetURL(id string) (string, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	v, ok := r.storage[id]
	if !ok {
		return "", fmt.Errorf("url not found by id: %v", id)
	}
	return v, nil
}

func NewRepository() *Repository {
	return &Repository{
		m:       sync.RWMutex{},
		storage: make(map[string]string),
	}
}
