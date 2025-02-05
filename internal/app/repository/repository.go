package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/DeneesK/short-url/internal/app/storage/memorystorage"
	"github.com/DeneesK/short-url/internal/app/storage/postgres"
)

// TODO CONF and OPTS RE-CHECK!

const filePerm = 0644

type StorageConfig struct {
	DBDSN           string
	MigrationSource string
	MaxStorageSize  uint64
}

type row struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

type Storage interface {
	Store(ctx context.Context, id, value string) error
	Get(ctx context.Context, id string) (string, error)
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
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
			conf.MaxStorageSize,
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
			err = rep.storage.Store(context.Background(), r.ShortURL, r.LongURL)
			if err != nil {
				return err
			}

		}
		return nil
	}
}

func (rep *Repository) Store(ctx context.Context, id, value string) error {
	if err := rep.storage.Store(ctx, id, value); err != nil {
		return err
	}
	if rep.encoder != nil {
		if err := rep.storeToFile(id, value); err != nil {
			return err
		}
	}
	return nil
}

func (rep *Repository) Get(ctx context.Context, id string) (string, error) {
	return rep.storage.Get(ctx, id)
}

func (rep *Repository) PingDB(ctx context.Context) error {
	return rep.storage.Ping(ctx)
}

func (rep *Repository) Close(ctx context.Context) error {
	err := rep.storage.Close(ctx)
	if err != nil {
		return err
	}
	return rep.file.Close()
}

func (rep *Repository) storeToFile(id, value string) error {
	r := row{ShortURL: id, LongURL: value}
	return rep.encoder.Encode(r)
}
