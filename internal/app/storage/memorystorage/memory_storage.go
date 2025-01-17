package memorystorage

import (
	"sync"

	"github.com/DeneesK/short-url/internal/app/storage"
)

const stringOverhead = 16

type MemoryStorage struct {
	m                sync.RWMutex
	storage          map[string]string
	currentBytesSize uint64
	maxStorageSize   uint64
}

func (s *MemoryStorage) Store(id, value string) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.isExists(id) {
		return storage.ErrNotUniqueID
	} else if s.currentBytesSize > s.maxStorageSize {
		return storage.ErrStorageLimitExceeded
	}

	s.storage[id] = value
	s.updateSize(id, value)
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

func (s *MemoryStorage) isExists(id string) bool {
	_, e := s.storage[id]
	return e
}

func (s *MemoryStorage) updateSize(newRow ...string) {
	var size uint64
	for _, s := range newRow {
		size += uint64(len(s)) + stringOverhead
	}
	s.currentBytesSize += size
}

func NewMemoryStorage(maxStorageSize uint64) *MemoryStorage {
	return &MemoryStorage{
		m:              sync.RWMutex{},
		storage:        make(map[string]string),
		maxStorageSize: maxStorageSize,
	}
}
