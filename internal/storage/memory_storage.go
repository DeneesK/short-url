package storage

import (
	"sync"
)

const stringOverhead = 16

type MemoryStorage struct {
	m                sync.RWMutex
	storage          map[string]string
	currentBytesSize uint64
}

func (s *MemoryStorage) Save(id, value string) error {
	if s.isExists(id) {
		return ErrNotUniqueID
	}
	s.m.Lock()
	s.storage[id] = value
	s.updateSize(id, value)
	s.m.Unlock()
	return nil
}

func (s *MemoryStorage) Get(id string) (string, error) {
	s.m.RLock()
	v, ok := s.storage[id]
	s.m.RUnlock()
	if !ok {
		return "", nil
	}
	return v, nil
}

func (s *MemoryStorage) CurrentSize() uint64 {
	return s.currentBytesSize
}

func (s *MemoryStorage) isExists(id string) bool {
	s.m.RLock()
	_, e := s.storage[id]
	s.m.RUnlock()
	return e
}

func (s *MemoryStorage) updateSize(newRow ...string) {
	var size uint64
	s.m.Lock()
	for _, s := range newRow {
		size += uint64(len(s)) + stringOverhead
	}
	s.currentBytesSize += size
	s.m.Unlock()
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		m:       sync.RWMutex{},
		storage: make(map[string]string),
	}
}
