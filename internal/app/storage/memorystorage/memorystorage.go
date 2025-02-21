package memorystorage

import (
	"context"
	"sync"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/storage"
	"github.com/google/uuid"
)

const stringOverhead = 16

type MemoryStorage struct {
	m                     sync.RWMutex
	storage               map[string]string
	uniqueValueConstraint map[string]string
	userIDreference       map[string][]string
	deletedURLs           map[string]bool
	currentBytesSize      uint64
	maxStorageSize        uint64
}

func NewMemoryStorage(maxStorageSize uint64) *MemoryStorage {
	return &MemoryStorage{
		storage:               make(map[string]string),
		uniqueValueConstraint: make(map[string]string),
		userIDreference:       make(map[string][]string),
		deletedURLs:           make(map[string]bool),
		maxStorageSize:        maxStorageSize,
	}
}

func (s *MemoryStorage) Store(ctx context.Context, id, value, userID string) (string, error) {
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
	s.userIDreference[userID] = append(s.userIDreference[userID], id)
	s.updateSize(id, value)
	return id, nil
}

func (s *MemoryStorage) StoreBatch(ctx context.Context, batch []dto.OriginalURL, userID string) error {
	for _, entity := range batch {
		_, err := s.Store(ctx, entity.ID, entity.URL, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, id string) (dto.LongUrl, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	longUrl := s.storage[id]
	isDeleted := s.deletedURLs[id]
	return dto.LongUrl{LongURL: longUrl, IsDeleted: isDeleted}, nil
}

func (s *MemoryStorage) GetByUserID(ctx context.Context, userID string) ([]dto.OriginalURL, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	urls := make([]dto.OriginalURL, 0)
	ids := s.userIDreference[userID]
	for _, id := range ids {
		originalURL, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		urls = append(urls, dto.OriginalURL{ID: id, URL: originalURL.LongURL})
	}
	return urls, nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) Close(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) CreateUser(ctx context.Context) (string, error) {
	newUUID := uuid.New()
	userID := newUUID.String()
	return userID, nil
}

func (s *MemoryStorage) UpdateStatusBatch(batch []dto.UpdateTask) error {
	s.m.Lock()
	defer s.m.Unlock()
	for _, task := range batch {
		ids := s.userIDreference[task.UserID]
		for _, id := range ids {
			if id == task.ID {
				s.deletedURLs[id] = true
				break
			}
		}
	}
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
