package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"sample-observability-api/internal/buildinfo"
	"sample-observability-api/internal/config"
	"sample-observability-api/internal/httpapi"
	"sample-observability-api/internal/observability"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	state := config.NewState()
	metrics := observability.NewRecorder()
	server := httpapi.NewServer(httpapi.ServerOptions{
		State:   state,
		Metrics: metrics,
		Logger:  logger,
		Build:   buildinfo.Current(),
	})

	addr := ":" + env("PORT", "8080")
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting sample observability api", "addr", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}