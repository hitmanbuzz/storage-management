package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"storage-management/internal/database"
	"storage-management/internal/server"
	"time"

	"golang.org/x/sync/errgroup"
)

type appState struct {
	server *server.Server
	db     *database.DatabaseHandler
	logger *slog.Logger
}

func NewAppState(ctx context.Context, logger *slog.Logger) *appState {
	serverIP := os.Getenv("SERVER_IP")
	dbURL := os.Getenv("DB_URL")

	db, err := database.NewDatabaseHandler(ctx, logger, dbURL)
	if err != nil {
		logger.Error("error", "failed to initialized database", err)
		return nil
	}

	return &appState{
		server: server.NewServer(serverIP, logger),
		db:     db,
		logger: logger,
	}
}

func (a *appState) Run(pctx context.Context) error {
	g, ctx := errgroup.WithContext(pctx)

	// handle server
	g.Go(func() error {
		if err := a.server.Run(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	// handle database
	g.Go(func() error {
		if err := a.db.Run(ctx); err != nil {
			return fmt.Errorf("database crashed: %w", err)
		}
		return nil
	})

	// handle server shutdown
	g.Go(func() error {
		<-ctx.Done()

		a.logger.Info("server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.server.GetHttpServer().Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server forced to shutdown with error: %w", err)
		}
		return nil
	})

	return g.Wait()
}
