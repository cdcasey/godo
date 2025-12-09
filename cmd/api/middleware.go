package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func loggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				logger.Info("request completed",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration", time.Since(start),
					"request_id", middleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{allowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

func authRateLimiter(logger *slog.Logger) func(http.Handler) http.Handler {
	return httprate.Limit(
		5,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			logger.Warn("rate limit exceeded",
				"ip", r.RemoteAddr,
				"path", r.URL.Path,
				"request_id", middleware.GetReqID(r.Context()),
			)
			http.Error(w, `{"error": "too many requests"}`, http.StatusTooManyRequests)
		}),
	)
}
