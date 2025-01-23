package repository

import (
	"bufio"
	"encoding/json"
	"os"
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
}

type Option func(*Repository) error

func NewRepository(storage Storage, opts ...Option) (*Repository, error) {
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

func WithDumpFile(dumpFilePath string) Option {
	return func(rep *Repository) error {
		file, err := os.OpenFile(dumpFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, filePerm)
		if err != nil {
			return err
		}
		rep.file = file
		rep.encoder = json.NewEncoder(file)
		return nil
	}
}

func RestoreFromDump(dumpFilePath string) Option {
	return func(rep *Repository) error {
		scanner := bufio.NewScanner(rep.file)
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
	if err := rep.storeToFile(id, value); err != nil {
		return err
	}
	return nil
}

func (rep *Repository) Get(id string) (string, error) {
	return rep.storage.Get(id)
}

func (rep *Repository) Close() error {
	return rep.file.Close()
}

func (rep *Repository) storeToFile(id, value string) error {
	r := row{ShortURL: id, LongURL: value}
	return rep.encoder.Encode(r)
}
