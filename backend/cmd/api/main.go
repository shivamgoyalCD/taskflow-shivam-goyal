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
	"taskflow-shivam-goyal/backend/internal/realtime"
	"taskflow-shivam-goyal/backend/internal/response"
	"taskflow-shivam-goyal/backend/internal/tasks"
)

type application struct {
	logger          *slog.Logger
	authHandler     *auth.Handler
	projectsHandler *projects.Handler
	tasksHandler    *tasks.Handler
	jwtManager      *auth.JWTManager
}

type healthResponse struct {
	Status string `json:"status"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("application_starting")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	startupCtx, cancelStartup := db.StartupContext(context.Background())
	defer cancelStartup()

	logger.Info(
		"db_connect_starting",
		"host", cfg.Postgres.Host,
		"port", cfg.Postgres.Port,
		"database", cfg.Postgres.Database,
	)

	pool, err := db.Open(startupCtx, cfg.Postgres)
	if err != nil {
		logger.Error("db_connect_failed", "error", err)
		os.Exit(1)
	}

	logger.Info("db_connected")

	if err := db.RunMigrations(startupCtx, logger, cfg.Postgres, pool); err != nil {
		logger.Error("db_migrations_failed", "error", err)
		closeDBPool(logger, pool)
		os.Exit(1)
	}

	logger.Info("db_seed_starting")

	if err := db.RunSeed(startupCtx, logger, pool); err != nil {
		logger.Error("db_seed_failed", "error", err)
		closeDBPool(logger, pool)
		os.Exit(1)
	}

	authRepository := auth.NewRepository(pool)
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry)
	realtimeManager := realtime.NewManager()
	authService := auth.NewService(authRepository, jwtManager)
	authHandler := auth.NewHandler(logger, authService)
	projectsRepository := projects.NewRepository(pool)
	projectsService := projects.NewService(projectsRepository)
	projectsHandler := projects.NewHandler(logger, projectsService, realtimeManager)
	tasksRepository := tasks.NewRepository(pool)
	tasksService := tasks.NewService(tasksRepository, realtimeManager)
	tasksHandler := tasks.NewHandler(logger, tasksService)

	app := &application{
		logger:          logger,
		authHandler:     authHandler,
		projectsHandler: projectsHandler,
		tasksHandler:    tasksHandler,
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

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdownSignals)

	select {
	case err := <-serverErrors:
		logger.Error("http_server_failed", "error", err)
		closeDBPool(logger, pool)
		os.Exit(1)
	case sig := <-shutdownSignals:
		logger.Info("shutdown_requested", "signal", sig.String())
	}

	logger.Info("http_server_shutdown_starting")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http_server_shutdown_failed", "error", err)
		closeDBPool(logger, pool)
		os.Exit(1)
	}

	logger.Info("http_server_stopped")
	closeDBPool(logger, pool)
	logger.Info("application_stopped")
}

func newRouter(app *application) http.Handler {
	router := chi.NewRouter()

	router.Use(appmiddleware.CORS())
	router.Use(chimiddleware.RequestID)
	router.Use(appmiddleware.RequestLogger(app.logger))
	router.Use(appmiddleware.Recoverer(app.logger))

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if err := response.NotFound(w, "resource not found"); err != nil {
			app.logger.Error("http_not_found_response_failed", "error", err)
		}
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		if err := response.MethodNotAllowed(w); err != nil {
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
			if err := response.OK(w, healthResponse{Status: "ok"}); err != nil {
				app.logger.Error("http_health_response_failed", "error", err)
			}
		})

		r.Get("/users", app.authHandler.ListUsers)
		r.Get("/projects", app.projectsHandler.List)
		r.Post("/projects", app.projectsHandler.Create)
		r.Get("/projects/{id}", app.projectsHandler.GetByID)
		r.Get("/projects/{id}/stats", app.projectsHandler.GetStats)
		r.Get("/projects/{id}/events", app.projectsHandler.Events)
		r.Patch("/projects/{id}", app.projectsHandler.Update)
		r.Delete("/projects/{id}", app.projectsHandler.Delete)
		r.Get("/projects/{id}/tasks", app.tasksHandler.ListByProject)
		r.Post("/projects/{id}/tasks", app.tasksHandler.Create)
		r.Patch("/tasks/{id}", app.tasksHandler.Update)
		r.Delete("/tasks/{id}", app.tasksHandler.Delete)

	})

	return router
}

func closeDBPool(logger *slog.Logger, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}

	logger.Info("db_pool_closing")
	pool.Close()
	logger.Info("db_pool_closed")
}
