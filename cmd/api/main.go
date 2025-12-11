package main

import (
	"godo/internal/auth"
	"godo/internal/config"
	"godo/internal/handlers"
	"godo/internal/service"
	"godo/internal/store"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := setupLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting todo API server")

	db, err := store.NewDB(cfg.DatabaseURL, cfg.DatabaseAuthToken)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := store.RunMigrations(db, "./migrations"); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Database migrations completed")

	// Repositories
	userRepo := store.NewUserRepo(db)
	todoRepo := store.NewTodoRepo(db)

	// Services
	authService := service.NewAuthService(userRepo)
	todoService := service.NewTodoService(todoRepo)
	userService := service.NewUserService(userRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, logger, cfg.JWTSecret)
	todoHandler := handlers.NewTodoHandler(todoService, logger)
	userHandler := handlers.NewUserHandler(userService, logger)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(loggerMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware(cfg.AllowedOrigins))

	r.Get("/api/health", healthHandler())

	r.With(authRateLimiter(logger)).Post("/api/register", authHandler.Register)
	r.With(authRateLimiter(logger)).Post("/api/login", authHandler.Login)

	r.Route("/api/todos", func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))
		r.Post("/", todoHandler.Create)
		r.Get("/", todoHandler.List)
		r.Get("/{id}", todoHandler.GetByID)
		r.Patch("/{id}", todoHandler.Update)
		r.Delete("/{id}", todoHandler.Delete)
	})

	r.Route("/api/users", func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))
		r.Get("/", userHandler.List)
		r.Get("/{id}", userHandler.GetByID)
		r.Patch("/{id}", userHandler.Update)
		r.Delete("/{id}", userHandler.Delete)
	})

	addr := ":" + cfg.Port
	logger.Info("Server starting", "port", cfg.Port)

	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
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
