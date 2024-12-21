package storage

import (
	"fmt"
	"sync"
)

type MemoryStorage struct {
	m       sync.RWMutex
	storage map[string]string
}

func (r *MemoryStorage) Save(id, value string) (string, error) {
	r.m.Lock()
	defer r.m.Unlock()
	r.storage[id] = value
	return id, nil
}

func (r *MemoryStorage) Get(id string) (string, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	v, ok := r.storage[id]
	if !ok {
		return "", fmt.Errorf("url not found by id: %v", id)
	}
	return v, nil
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		m:       sync.RWMutex{},
		storage: make(map[string]string),
	}
}
