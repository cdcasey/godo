package main

import (
	"godo/internal/config"
	"godo/internal/store"
	"log"
	"log/slog"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := setupLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting todo API server")

	db, err := store.New(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Database migrations completed")

}

func setupLogger(level, format string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: logLevel}

	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
