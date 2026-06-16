package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"storage-management/internal"
	"syscall"

	"github.com/joho/godotenv"
)

const (
	IP_ADDR = "0.0.0.0:6969"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error reading env file:", err)
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	app := internal.NewAppState(ctx, logger)
	if app == nil {
		logger.Error("failed to initialized app state")
		return // have error log already during function called
	}

	if err = app.Run(ctx); err != nil {
		logger.Error("application force shutdown", "error", err)
	} else {
		logger.Debug("application shutdown gracefully")
	}
}
