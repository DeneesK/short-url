package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/storage"
	"github.com/DeneesK/short-url/internal/app/storage/memorystorage"
	"github.com/DeneesK/short-url/internal/app/storage/postgres"
)

const filePerm = 0644

type StorageConfig struct {
	DBDSN           string
	MigrationSource string
	MaxStorageSize  uint64
}

type row struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
	UserID   string `json:"user_id,omitempty"`
}

type Storage interface {
	Store(ctx context.Context, id, value, userID string) (string, error)
	StoreBatch(ctx context.Context, batch []dto.OriginalURL, userID string) error
	UpdateStatusBatch([]dto.UpdateTask) error
	Get(context.Context, string) (dto.LongURL, error)
	GetByUserID(ctx context.Context, userID string) ([]dto.OriginalURL, error)
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context) (string, error)
}

type Repository struct {
	storage Storage
	file    *os.File
	encoder *json.Encoder
}

type Option func(*Repository) error

func NewRepository(conf StorageConfig, opts ...Option) (*Repository, error) {
	var storage Storage
	if conf.DBDSN != "" {
		ctx := context.Background()
		storage = postgres.NewDBConnection(
			ctx, conf.DBDSN,
			postgres.RunMigrations(conf.MigrationSource, conf.DBDSN),
		)
	} else {
		storage = memorystorage.NewMemoryStorage(conf.MaxStorageSize)
	}
	rep := &Repository{
		storage: storage,
	}

	for _, opt := range opts {
		if err := opt(rep); err != nil {
			return nil, err
		}
	}

	return rep, nil
}

func AddDumpFile(dumpFilePath string) Option {
	return func(rep *Repository) error {
		if dumpFilePath == "" {
			return nil
		}
		file, err := os.OpenFile(dumpFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePerm)
		if err != nil {
			return err
		}
		rep.file = file
		rep.encoder = json.NewEncoder(rep.file)
		return nil
	}
}

func RestoreFromDump(dumpFilePath string) Option {
	return func(rep *Repository) error {
		if dumpFilePath == "" {
			return nil
		}
		file, err := os.OpenFile(dumpFilePath, os.O_RDONLY, filePerm)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			data := scanner.Bytes()

			r := row{}
			err := json.Unmarshal(data, &r)
			if err != nil {
				return err
			}
			_, err = rep.storage.Store(context.Background(), r.ShortURL, r.LongURL, r.UserID)
			if err != nil {
				return err
			}

		}
		return nil
	}
}

func (rep *Repository) Store(ctx context.Context, id, value, userID string) (string, error) {
	if alias, err := rep.storage.Store(ctx, id, value, userID); err != nil && err != storage.ErrUniqueViolation {
		return "", err
	} else if errors.Is(err, storage.ErrUniqueViolation) {
		return alias, storage.ErrUniqueViolation
	}
	if rep.encoder != nil {
		if err := rep.storeToFile(id, value, userID); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (rep *Repository) StoreBatch(ctx context.Context, batch []dto.OriginalURL, userID string) error {

	err := rep.storage.StoreBatch(ctx, batch, userID)
	if err != nil {
		return err
	}

	if rep.encoder != nil {
		for _, entry := range batch {
			if err := rep.storeToFile(entry.ID, entry.URL, userID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (rep *Repository) UpdateStatusBatch(batch []dto.UpdateTask) error {

	return rep.storage.UpdateStatusBatch(batch)
}

func (rep *Repository) Get(ctx context.Context, id string) (dto.LongURL, error) {
	return rep.storage.Get(ctx, id)
}

func (rep *Repository) GetByUserID(ctx context.Context, userID string) ([]dto.OriginalURL, error) {
	return rep.storage.GetByUserID(ctx, userID)
}

func (rep *Repository) PingDB(ctx context.Context) error {
	return rep.storage.Ping(ctx)
}

func (rep *Repository) Close(ctx context.Context) error {
	err := rep.storage.Close(ctx)
	if err != nil {
		return err
	}
	if rep.file != nil {
		return rep.file.Close()
	}
	return nil
}

func (rep *Repository) CreateUser(ctx context.Context) (string, error) {
	return rep.storage.CreateUser(ctx)
}

func (rep *Repository) storeToFile(id, value, userID string) error {
	r := row{ShortURL: id, LongURL: value, UserID: userID}
	return rep.encoder.Encode(r)
}
