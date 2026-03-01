package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"cloud-webdav-server/internal/config"
	"cloud-webdav-server/internal/server"
)

func main() {
	// Structured JSON logging in production; text for local dev.
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(handler))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg)
	if err != nil {
		slog.Error("server init", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.Start(ctx); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
