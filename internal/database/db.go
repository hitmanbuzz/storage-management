package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseHandler struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewDatabaseHandler(ctx context.Context, logger *slog.Logger, connURL string) (*DatabaseHandler, error) {
	pool, err := pgxpool.New(ctx, connURL)
	if err != nil {
		return nil, err
	}

	return &DatabaseHandler{
		pool: pool,
	}, nil
}
