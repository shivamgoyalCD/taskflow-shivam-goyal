package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"taskflow-shivam-goyal/backend/internal/config"
)

const (
	defaultSchemaName       = "public"
	defaultMigrationsTable  = "schema_migrations"
	defaultConnectTimeout   = 5 * time.Second
	defaultMigrationTimeout = 30 * time.Second
	seedUserID              = "11111111-1111-4111-8111-111111111111"
	seedProjectID           = "22222222-2222-4222-8222-222222222222"
	seedTaskTodoID          = "33333333-3333-4333-8333-333333333331"
	seedTaskInProgressID    = "33333333-3333-4333-8333-333333333332"
	seedTaskDoneID          = "33333333-3333-4333-8333-333333333333"
)

type TableCounts struct {
	Users    int64 `json:"users"`
	Projects int64 `json:"projects"`
	Tasks    int64 `json:"tasks"`
}

func Open(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(connectionString(cfg))
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	if poolConfig.ConnConfig.ConnectTimeout == 0 {
		poolConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("open postgres pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func RunMigrations(ctx context.Context, logger *slog.Logger, cfg config.PostgresConfig, pool *pgxpool.Pool) error {
	migrationsDir, err := resolveMigrationsDir()
	if err != nil {
		return err
	}

	logger.Info("db_migrations_starting", "source", migrationsDir)

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	driver, err := pgxmigrate.WithInstance(sqlDB, &pgxmigrate.Config{
		DatabaseName:          cfg.Database,
		SchemaName:            defaultSchemaName,
		MigrationsTable:       defaultMigrationsTable,
		MultiStatementEnabled: true,
	})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	sourceURL := migrationSourceURL(migrationsDir)
	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, "pgx/v5", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	migrator.Log = migrateLogger{logger: logger}

	defer func() {
		sourceErr, databaseErr := migrator.Close()
		if sourceErr != nil {
			logger.Error("db_migration_source_close_failed", "error", sourceErr)
		}
		if databaseErr != nil {
			logger.Error("db_migration_driver_close_failed", "error", databaseErr)
		}
	}()

	done := make(chan error, 1)

	go func() {
		done <- migrator.Up()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("run migrations: %w", ctx.Err())
	case err := <-done:
		if err == nil {
			logger.Info("db_migrations_applied")
			return nil
		}

		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("db_migrations_no_change")
			return nil
		}

		return fmt.Errorf("run migrations: %w", err)
	}
}

func RunSeed(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) error {
	seedFile, err := resolveSeedFile()
	if err != nil {
		return err
	}

	applied, err := seedAlreadyApplied(ctx, pool)
	if err != nil {
		return fmt.Errorf("check seed state: %w", err)
	}
	if applied {
		logger.Info("db_seed_skipped", "reason", "already_applied", "source", seedFile)
		return nil
	}

	seedSQL, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("read seed file %q: %w", seedFile, err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire postgres connection for seed: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Conn().PgConn().Exec(ctx, string(seedSQL)).ReadAll(); err != nil {
		return fmt.Errorf("execute seed file %q: %w", seedFile, err)
	}

	logger.Info("db_seed_applied", "source", seedFile)

	return nil
}

func CountCoreTables(ctx context.Context, pool *pgxpool.Pool) (TableCounts, error) {
	var counts TableCounts

	err := pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM projects),
			(SELECT COUNT(*) FROM tasks)
	`).Scan(&counts.Users, &counts.Projects, &counts.Tasks)
	if err != nil {
		return TableCounts{}, fmt.Errorf("count core tables: %w", err)
	}

	return counts, nil
}

func StartupContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, defaultMigrationTimeout)
}

func connectionString(cfg config.PostgresConfig) string {
	return databaseURL("postgres", cfg)
}

func databaseURL(scheme string, cfg config.PostgresConfig) string {
	databaseURL := &url.URL{
		Scheme: scheme,
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.Database,
	}

	query := databaseURL.Query()
	query.Set("application_name", "taskflow-api")
	databaseURL.RawQuery = query.Encode()

	return databaseURL.String()
}

func resolveMigrationsDir() (string, error) {
	candidates := []string{
		"migrations",
		filepath.Join("backend", "migrations"),
		filepath.Join("..", "migrations"),
		filepath.Join("..", "backend", "migrations"),
	}

	if executablePath, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executablePath)
		candidates = append(candidates,
			filepath.Join(executableDir, "migrations"),
			filepath.Join(executableDir, "..", "..", "migrations"),
			filepath.Join(executableDir, "..", "..", "..", "migrations"),
		)
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			continue
		}

		absolutePath, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("resolve migrations path %q: %w", candidate, err)
		}

		return absolutePath, nil
	}

	return "", fmt.Errorf("migrations directory not found")
}

func resolveSeedFile() (string, error) {
	candidates := []string{
		filepath.Join("seeds", "seed.sql"),
		filepath.Join("backend", "seeds", "seed.sql"),
		filepath.Join("..", "seeds", "seed.sql"),
		filepath.Join("..", "backend", "seeds", "seed.sql"),
	}

	if executablePath, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executablePath)
		candidates = append(candidates,
			filepath.Join(executableDir, "seeds", "seed.sql"),
			filepath.Join(executableDir, "..", "..", "seeds", "seed.sql"),
			filepath.Join(executableDir, "..", "..", "..", "seeds", "seed.sql"),
		)
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}

		absolutePath, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("resolve seed path %q: %w", candidate, err)
		}

		return absolutePath, nil
	}

	return "", fmt.Errorf("seed file not found")
}

func migrationSourceURL(path string) string {
	return "file://" + filepath.ToSlash(path)
}

func seedAlreadyApplied(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	var (
		userExists           bool
		projectExists        bool
		todoTaskExists       bool
		inProgressTaskExists bool
		doneTaskExists       bool
	)

	err := pool.QueryRow(ctx, `
		SELECT
			EXISTS(SELECT 1 FROM users WHERE id = $1),
			EXISTS(SELECT 1 FROM projects WHERE id = $2),
			EXISTS(SELECT 1 FROM tasks WHERE id = $3),
			EXISTS(SELECT 1 FROM tasks WHERE id = $4),
			EXISTS(SELECT 1 FROM tasks WHERE id = $5)
	`, seedUserID, seedProjectID, seedTaskTodoID, seedTaskInProgressID, seedTaskDoneID).Scan(
		&userExists,
		&projectExists,
		&todoTaskExists,
		&inProgressTaskExists,
		&doneTaskExists,
	)
	if err != nil {
		return false, fmt.Errorf("query seed state: %w", err)
	}

	return userExists && projectExists && todoTaskExists && inProgressTaskExists && doneTaskExists, nil
}

type migrateLogger struct {
	logger *slog.Logger
}

func (l migrateLogger) Printf(format string, values ...any) {
	l.logger.Info("db_migration_log", "message", fmt.Sprintf(format, values...))
}

func (l migrateLogger) Verbose() bool {
	return true
}
