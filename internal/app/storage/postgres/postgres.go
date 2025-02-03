package postgres

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewConnection(ctx context.Context, dbDSN string) *sql.DB {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	return db
}
