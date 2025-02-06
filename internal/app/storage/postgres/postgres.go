package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/DeneesK/short-url/internal/app/storage"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	db             *sql.DB
	maxStorageSize uint64
}

type Option func(*PostgresStorage) error

func NewDBConnection(ctx context.Context, dbDSN string, maxStorageSize uint64, opts ...Option) *PostgresStorage {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	s := &PostgresStorage{db: db, maxStorageSize: maxStorageSize}
	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			log.Fatalf("Unable to apply option: %v", err)
		}
	}
	return s
}

func RunMigrations(migrationSource, dbDSN string) Option {
	return func(rep *PostgresStorage) error {
		m, err := migrate.New(
			migrationSource,
			dbDSN)
		if err != nil {
			return err
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
		log.Println("Migrations done")
		return nil
	}
}

func (s *PostgresStorage) Store(ctx context.Context, id, value string) (string, error) {
	currentSize, err := s.getDBSize(ctx)
	if err != nil {
		return "", err
	}
	if currentSize > s.maxStorageSize {
		return "", storage.ErrStorageLimitExceeded
	}

	var existingID string
	checkQuery := "SELECT alias FROM shorten_url WHERE long_url = $1"
	err = s.db.QueryRowContext(ctx, checkQuery, value).Scan(&existingID)
	if err == nil {
		return existingID, storage.ErrUniqueViolation
	} else if err != sql.ErrNoRows {
		return "", err
	}

	query := "INSERT INTO shorten_url (alias, long_url) VALUES ($1, $2)"
	_, err = s.db.ExecContext(ctx, query, id, value)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *PostgresStorage) StoreBatch(ctx context.Context, batch [][2]string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO shorten_url (alias, long_url) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	for _, entity := range batch {
		_, err := stmt.ExecContext(ctx, entity[0], entity[1])
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *PostgresStorage) Get(ctx context.Context, id string) (string, error) {
	query := "SELECT long_url FROM shorten_url WHERE alias = $1"
	var longURL string
	err := s.db.QueryRowContext(ctx, query, id).Scan(&longURL)
	if err != nil {
		return "", err
	}
	return longURL, nil
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.Ping()
}

func (s *PostgresStorage) Close(ctx context.Context) error {
	return s.db.Close()
}

func (s *PostgresStorage) getDBSize(ctx context.Context) (uint64, error) {
	var size uint64
	query := "SELECT pg_database_size(current_database());"

	err := s.db.QueryRowContext(ctx, query).Scan(&size)
	if err != nil {
		return 0, err
	}

	return size, nil
}
