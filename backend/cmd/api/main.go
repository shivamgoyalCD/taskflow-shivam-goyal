package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"taskflow-shivam-goyal/backend/internal/config"
	appmiddleware "taskflow-shivam-goyal/backend/internal/middleware"
	"taskflow-shivam-goyal/backend/internal/response"
)

type healthResponse struct {
	Status string `json:"status"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	router := newRouter(logger)

	server := &http.Server{
		Addr:              cfg.Server.Address(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("http_server_starting", "addr", server.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		logger.Error("http_server_failed", "error", err)
		os.Exit(1)
	case <-stopCtx.Done():
		logger.Info("http_server_shutdown_requested")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http_server_shutdown_failed", "error", err)
		os.Exit(1)
	}

	logger.Info("http_server_stopped")
}

func newRouter(logger *slog.Logger) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.RequestID)
	router.Use(appmiddleware.RequestLogger(logger))
	router.Use(appmiddleware.Recoverer(logger))

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if err := response.Error(w, http.StatusNotFound, "resource not found"); err != nil {
			logger.Error("http_not_found_response_failed", "error", err)
		}
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		if err := response.Error(w, http.StatusMethodNotAllowed, "method not allowed"); err != nil {
			logger.Error("http_method_not_allowed_response_failed", "error", err)
		}
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := response.JSON(w, http.StatusOK, healthResponse{Status: "ok"}); err != nil {
			logger.Error("http_health_response_failed", "error", err)
		}
	})

	return router
}
