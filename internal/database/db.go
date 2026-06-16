package database

import (
	"context"
	"fmt"
	"log/slog"
	"storage-management/internal/util"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseHandler struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
	dbURL  string
}

func NewDatabaseHandler(ctx context.Context, logger *slog.Logger, connURL string) (*DatabaseHandler, error) {
	return &DatabaseHandler{
		pool:   nil,
		logger: logger,
		dbURL:  connURL,
	}, nil
}

func (d *DatabaseHandler) connect(ctx context.Context) error {
	if d.pool != nil {
		return nil
	}

	pool, err := pgxpool.New(ctx, d.dbURL)
	if err != nil {
		return err
	}

	d.pool = pool
	if err = d.CheckHealth(ctx, util.MAX_DB_PING); err != nil {
		d.logger.Error("database connection", "status", "failed", "error", err)
		return err
	}

	d.logger.Debug("database connection", "status", "success")
	return nil
}

func (d *DatabaseHandler) Run(ctx context.Context) error {
	if err := d.connect(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("database shutdown initiated")
			closed := make(chan struct{})
			go func() {
				d.pool.Close()
				close(closed)
			}()

			select {
			case <-closed:
				d.logger.Info("database connection pool closed gracefully")
			case <-time.After(5 * time.Second):
				d.logger.Warn("database connection pool forced to close due to timeout")
			}
			return nil

		case <-ticker.C:
			if err := d.CheckHealth(ctx, util.MAX_DB_PING); err != nil {
				return fmt.Errorf("database health check failed after %d attempts: %w", util.MAX_DB_PING, err)
			}
			d.logger.Debug("database ping successful")
		}
	}
}

func (d *DatabaseHandler) CheckHealth(ctx context.Context, maxAttempts int) error {
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = d.pool.Ping(ctx)
		if err == nil {
			return nil
		}

		d.logger.Warn("database ping failed", "attempt", attempt, "error", err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	return err
}
