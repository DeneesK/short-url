package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/DeneesK/short-url/internal/app/storage/postgres"
	"github.com/jackc/pgx/v5"
)

const filePerm = 0644

type row struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

type Storage interface {
	Store(id, value string) error
	Get(id string) (string, error)
}

type Repository struct {
	storage Storage
	file    *os.File
	encoder *json.Encoder
	conn    *pgx.Conn
}

type Option func(*Repository) error

func NewRepository(storage Storage, opts ...Option) (*Repository, error) {
	conn := postgres.NewConnection(context.TODO(), os.Getenv("DATABASE_DSN"))
	rep := &Repository{
		storage: storage,
		conn:    conn,
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

			err = rep.storage.Store(r.ShortURL, r.LongURL)
			if err != nil {
				return err
			}

		}
		return nil
	}
}

func (rep *Repository) Store(id, value string) error {
	if err := rep.storage.Store(id, value); err != nil {
		return err
	}
	if rep.encoder != nil {
		if err := rep.storeToFile(id, value); err != nil {
			return err
		}
	}
	return nil
}

func (rep *Repository) Get(id string) (string, error) {
	return rep.storage.Get(id)
}

func (rep *Repository) PingDB() error {
	return rep.conn.Ping(context.TODO())
}

func (rep *Repository) Close() error {
	err := rep.conn.Close(context.TODO())
	if err != nil {
		return err
	}
	return rep.file.Close()
}

func (rep *Repository) storeToFile(id, value string) error {
	r := row{ShortURL: id, LongURL: value}
	return rep.encoder.Encode(r)
}
