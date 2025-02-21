package postgres

import (
	"context"
	"database/sql"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/DeneesK/short-url/internal/app/dto"
	"github.com/DeneesK/short-url/internal/app/storage"
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

func (s *PostgresStorage) Store(ctx context.Context, id, value, userID string) (string, error) {
	query := "INSERT INTO shorten_url (alias, long_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (long_url) DO UPDATE SET alias = shorten_url.alias RETURNING alias"
	var alias string

	err := s.db.QueryRowContext(ctx, query, id, value, userID).Scan(&alias)
	if err != nil {
		return "", err
	}

	if alias != id {
		return alias, storage.ErrUniqueViolation
	}

	return id, nil
}

func (s *PostgresStorage) StoreBatch(ctx context.Context, batch []dto.OriginalURL, userID string) error {
	const chunkSize = 1000
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

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

		psql := psql.Insert("shorten_url").Columns("alias", "long_url", "user_id")

		for _, row := range chunk {
			psql = psql.Values(row.ID, row.URL, userID)
		}

		query, args, err := psql.ToSql()
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStorage) Get(ctx context.Context, id string) (dto.LongURL, error) {
	query := "SELECT long_url, is_deleted FROM shorten_url WHERE alias = $1"
	var longURL dto.LongURL
	err := s.db.QueryRowContext(ctx, query, id).Scan(&longURL.LongURL, &longURL.IsDeleted)
	if err != nil {
		return dto.LongURL{}, err
	}
	return longURL, nil
}

func (s *PostgresStorage) GetByUserID(ctx context.Context, userID string) ([]dto.OriginalURL, error) {
	query := "SELECT long_url, alias FROM shorten_url WHERE user_id = $1 and is_deleted = false"

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	urls := make([]dto.OriginalURL, 0)
	for rows.Next() {
		var LongURL string
		var alias string
		err := rows.Scan(&LongURL, &alias)
		if err != nil {
			return nil, err
		}
		url := dto.OriginalURL{ID: alias, URL: LongURL}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

func (s *PostgresStorage) UpdateStatusBatch(batch []dto.UpdateTask) error {
	const chunkSize = 1000
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

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

		ids := make([]string, len(chunk))
		userIDs := make([]string, len(chunk))
		for i, row := range chunk {
			ids[i] = row.ID
			userIDs[i] = row.UserID
		}

		psql := psql.Update("shorten_url").
			Set("is_deleted", true).
			Where(sq.Eq{"id": ids, "user_id": userIDs})

		query, args, err := psql.ToSql()
		if err != nil {
			return err
		}
		_, err = tx.Exec(query, args...)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.Ping()
}

func (s *PostgresStorage) Close(ctx context.Context) error {
	return s.db.Close()
}

func (s *PostgresStorage) CreateUser(ctx context.Context) (string, error) {
	query := "INSERT INTO users (id) VALUES ($1)"
	newUUID := uuid.New()
	userID := newUUID.String()
	_, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}
