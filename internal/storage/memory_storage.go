package storage

import (
	"sync"
)

type MemoryStorage struct {
	m       sync.RWMutex
	storage map[string]string
}

func (r *MemoryStorage) Save(id, value string) error {
	if r.isExists(id) {
		return ErrNotUniqueID
	}
	r.m.Lock()
	r.storage[id] = value
	r.m.Unlock()
	return nil
}

func (r *MemoryStorage) Get(id string) (string, error) {
	r.m.RLock()
	v, ok := r.storage[id]
	r.m.RUnlock()
	if !ok {
		return "", nil
	}
	return v, nil
}

func (r *MemoryStorage) isExists(id string) bool {
	r.m.RLock()
	_, e := r.storage[id]
	r.m.RUnlock()
	return e
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		m:       sync.RWMutex{},
		storage: make(map[string]string),
	}
}
