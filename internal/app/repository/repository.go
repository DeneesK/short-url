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
	dumpingFilePath string
	storage         Storage
	encoder         *json.Encoder
}

func NewRepository(storage Storage, dumpingFilePath string) (*Repository, error) {
	rep := &Repository{
		dumpingFilePath: dumpingFilePath,
		storage:         storage,
	}
	file, err := os.OpenFile(dumpingFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, filePerm)
	if err != nil {
		return nil, err
	}
	rep.encoder = json.NewEncoder(file)
	err = rep.restoreFromDump(dumpingFilePath)
	if err != nil {
		return nil, err
	}
	return rep, nil
}

func (rep *Repository) Store(id, value string) error {
	err := rep.storage.Store(id, value)
	if err != nil {
		return err
	}
	err = rep.dumpToFile(id, value)
	if err != nil {
		return err
	}
	return nil
}

func (rep *Repository) Get(id string) (string, error) {
	return rep.storage.Get(id)
}

func (rep *Repository) dumpToFile(id, value string) error {
	return rep.encoder.Encode(row{ShortURL: id, LongURL: value})
}

func (rep *Repository) restoreFromDump(dumpingFilePath string) error {
	file, err := os.OpenFile(dumpingFilePath, os.O_RDONLY|os.O_CREATE, filePerm)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	var r row
	for scanner.Scan() {
		data := scanner.Bytes()

		err = json.Unmarshal(data, &r)
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
