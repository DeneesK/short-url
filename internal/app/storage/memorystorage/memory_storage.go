package memorystorage

import (
	"context"
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

func NewMemoryStorage(maxStorageSize uint64) *MemoryStorage {
	return &MemoryStorage{
		m:              sync.RWMutex{},
		storage:        make(map[string]string),
		maxStorageSize: maxStorageSize,
	}
}

func (s *MemoryStorage) Store(ctx context.Context, id, value string) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.isExists(id) {
		return "", storage.ErrNotUniqueID
	}

	for _id, v := range s.storage {
		if v == value {
			return _id, storage.ErrUniqueViolation
		}
	}

	if s.currentBytesSize > s.maxStorageSize {
		return "", storage.ErrStorageLimitExceeded
	}

	s.storage[id] = value
	s.updateSize(id, value)
	return id, nil
}

func (s *MemoryStorage) StoreBatch(ctx context.Context, batch [][2]string) error {
	for _, entity := range batch {
		_, err := s.Store(ctx, entity[0], entity[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, id string) (string, error) {
	s.m.RLock()
	v, ok := s.storage[id]
	s.m.RUnlock()
	if !ok {
		return "", nil
	}
	return v, nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) Close(ctx context.Context) error {
	return nil
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
