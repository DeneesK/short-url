package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/storage"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	db *sql.DB
}

type Option func(*PostgresStorage) error

func NewDBConnection(ctx context.Context, dbDSN string, opts ...Option) *PostgresStorage {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	s := &PostgresStorage{db: db}
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
	query := "INSERT INTO shorten_url (alias, long_url) VALUES ($1, $2) ON CONFLICT (long_url) DO UPDATE SET alias = shorten_url.alias RETURNING alias"
	var alias string

	err := s.db.QueryRowContext(ctx, query, id, value).Scan(&alias)
	if err != nil {
		return "", err
	}

	if alias != id {
		return alias, storage.ErrUniqueViolation
	}

	return id, nil
}

func (s *PostgresStorage) StoreBatch(ctx context.Context, batch []dto.OriginalURL) error {
	const chunkSize = 1000

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i := 0; i < len(batch); i += chunkSize {
		end := i + chunkSize
		if end > len(batch) {
			end = len(batch)
		}

		chunk := batch[i:end]
		var queryBuilder strings.Builder
		queryBuilder.WriteString("INSERT INTO shorten_url (alias, long_url) VALUES ")

		params := []interface{}{}
		for j, row := range chunk {
			if j > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(fmt.Sprintf("($%d, $%d)", 2*j+1, 2*j+2))
			params = append(params, row.ID, row.URL)
		}

		_, err = tx.ExecContext(ctx, queryBuilder.String(), params...)
		if err != nil {
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
