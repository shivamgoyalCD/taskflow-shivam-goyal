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
	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow-shivam-goyal/backend/internal/auth"
	"taskflow-shivam-goyal/backend/internal/config"
	"taskflow-shivam-goyal/backend/internal/db"
	appmiddleware "taskflow-shivam-goyal/backend/internal/middleware"
	"taskflow-shivam-goyal/backend/internal/projects"
	"taskflow-shivam-goyal/backend/internal/response"
)

type application struct {
	logger          *slog.Logger
	db              *pgxpool.Pool
	authHandler     *auth.Handler
	projectsHandler *projects.Handler
	jwtManager      *auth.JWTManager
}

type healthResponse struct {
	Status string `json:"status"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	startupCtx, cancelStartup := db.StartupContext(context.Background())
	defer cancelStartup()

	pool, err := db.Open(startupCtx, cfg.Postgres)
	if err != nil {
		logger.Error("db_connect_failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("db_connected")

	if err := db.RunMigrations(startupCtx, logger, cfg.Postgres, pool); err != nil {
		logger.Error("db_migrations_failed", "error", err)
		os.Exit(1)
	}

	if err := db.RunSeed(startupCtx, logger, pool); err != nil {
		logger.Error("db_seed_failed", "error", err)
		os.Exit(1)
	}

	authRepository := auth.NewRepository(pool)
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry)
	authService := auth.NewService(authRepository, jwtManager)
	authHandler := auth.NewHandler(logger, authService)
	projectsRepository := projects.NewRepository(pool)
	projectsService := projects.NewService(projectsRepository)
	projectsHandler := projects.NewHandler(logger, projectsService)

	app := &application{
		logger:          logger,
		db:              pool,
		authHandler:     authHandler,
		projectsHandler: projectsHandler,
		jwtManager:      jwtManager,
	}

	router := newRouter(app)

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

func newRouter(app *application) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.RequestID)
	router.Use(appmiddleware.RequestLogger(app.logger))
	router.Use(appmiddleware.Recoverer(app.logger))

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if err := response.Error(w, http.StatusNotFound, "resource not found"); err != nil {
			app.logger.Error("http_not_found_response_failed", "error", err)
		}
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		if err := response.Error(w, http.StatusMethodNotAllowed, "method not allowed"); err != nil {
			app.logger.Error("http_method_not_allowed_response_failed", "error", err)
		}
	})

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", app.authHandler.Register)
		r.Post("/login", app.authHandler.Login)
	})

	router.Group(func(r chi.Router) {
		r.Use(appmiddleware.Authenticate(app.logger, app.jwtManager))

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			if err := response.JSON(w, http.StatusOK, healthResponse{Status: "ok"}); err != nil {
				app.logger.Error("http_health_response_failed", "error", err)
			}
		})

		r.Get("/projects", app.projectsHandler.List)
		r.Post("/projects", app.projectsHandler.Create)
		r.Get("/projects/{id}", app.projectsHandler.GetByID)
		r.Patch("/projects/{id}", app.projectsHandler.Update)
		r.Delete("/projects/{id}", app.projectsHandler.Delete)

		// Temporary local debugging route. Remove before final submission.
		r.Get("/debug/seed-check", app.handleSeedCheck)
	})

	return router
}

func (app *application) handleSeedCheck(w http.ResponseWriter, r *http.Request) {
	counts, err := db.CountCoreTables(r.Context(), app.db)
	if err != nil {
		app.logger.Error("http_seed_check_failed", "error", err)
		if writeErr := response.Error(w, http.StatusInternalServerError, "failed to fetch seed counts"); writeErr != nil {
			app.logger.Error("http_seed_check_error_response_failed", "error", writeErr)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, counts); err != nil {
		app.logger.Error("http_seed_check_response_failed", "error", err)
	}
}
