package repository

import (
	"encoding/json"
	"log"
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

func NewRepository(storage Storage, dumpFilePath string) *Repository {
	file, err := os.OpenFile(dumpFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePerm)
	if err != nil {
		log.Fatalf("failed to initialized repository: %s", err)
	}

	return &Repository{
		storage: storage,
		file:    file,
		encoder: json.NewEncoder(file),
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
