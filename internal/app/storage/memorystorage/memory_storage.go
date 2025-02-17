package memorystorage

import (
	"context"
	"sync"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/storage"
)

const stringOverhead = 16

type MemoryStorage struct {
	m                     sync.RWMutex
	storage               map[string]string
	uniqueValueConstraint map[string]string
	currentBytesSize      uint64
	maxStorageSize        uint64
}

func NewMemoryStorage(maxStorageSize uint64) *MemoryStorage {
	return &MemoryStorage{
		storage:               make(map[string]string),
		uniqueValueConstraint: make(map[string]string),
		maxStorageSize:        maxStorageSize,
	}
}

func (s *MemoryStorage) Store(ctx context.Context, id, value string) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.isIDExists(id) {
		return "", storage.ErrNotUniqueID
	}

	if s.isValueExists(value) {
		return s.uniqueValueConstraint[value], storage.ErrUniqueViolation
	}

	if s.currentBytesSize > s.maxStorageSize {
		return "", storage.ErrStorageLimitExceeded
	}

	s.storage[id] = value
	s.uniqueValueConstraint[value] = id
	s.updateSize(id, value)
	return id, nil
}

func (s *MemoryStorage) StoreBatch(ctx context.Context, batch []dto.OriginalURL) error {
	for _, entity := range batch {
		_, err := s.Store(ctx, entity.ID, entity.URL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, id string) (string, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.storage[id], nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) Close(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) isIDExists(id string) bool {
	_, e := s.storage[id]
	return e
}

func (s *MemoryStorage) isValueExists(value string) bool {
	_, e := s.uniqueValueConstraint[value]
	return e
}

func (s *MemoryStorage) updateSize(newRow ...string) {
	var size uint64
	for _, s := range newRow {
		size += (uint64(len(s)) + stringOverhead) * 2
	}
	s.currentBytesSize += size
}
