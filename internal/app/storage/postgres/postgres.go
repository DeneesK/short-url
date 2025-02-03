package postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func NewConnection(ctx context.Context, dbURL string) *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	return conn
}
