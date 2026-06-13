package main

import (
	"log/slog"
	"os"
	"storage-management/internal/server"
)

const (
	IP_ADDR = "0.0.0.0:6969"
)

func main() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	server := server.NewServer(IP_ADDR, logger)
	server.Routes()
	server.Run()
}
