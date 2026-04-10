package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port int
}

type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

func Load() (Config, error) {
	if err := loadEnvFiles(
		".env",
		filepath.Join("..", ".env"),
	); err != nil {
		return Config{}, err
	}

	loader := envLoader{}

	cfg := Config{
		Server: ServerConfig{
			Port: loader.requiredPort("SERVER_PORT"),
		},
		Postgres: PostgresConfig{
			Host:     loader.requiredString("POSTGRES_HOST"),
			Port:     loader.requiredPort("POSTGRES_PORT"),
			Database: loader.requiredString("POSTGRES_DB"),
			User:     loader.requiredString("POSTGRES_USER"),
			Password: loader.requiredString("POSTGRES_PASSWORD"),
		},
		JWT: JWTConfig{
			Secret: loader.requiredString("JWT_SECRET"),
			Expiry: loader.requiredHours("JWT_EXPIRY_HOURS"),
		},
	}

	if err := loader.err(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c ServerConfig) Address() string {
	return fmt.Sprintf(":%d", c.Port)
}

type envLoader struct {
	errs []error
}

func (l *envLoader) requiredString(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		l.errs = append(l.errs, fmt.Errorf("%s is required", key))
		return ""
	}

	value = strings.TrimSpace(value)
	if value == "" {
		l.errs = append(l.errs, fmt.Errorf("%s cannot be empty", key))
		return ""
	}

	return value
}

func (l *envLoader) requiredPort(key string) int {
	value := l.requiredString(key)
	if value == "" {
		return 0
	}

	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		l.errs = append(l.errs, fmt.Errorf("%s must be a valid TCP port", key))
		return 0
	}

	return port
}

func (l *envLoader) requiredHours(key string) time.Duration {
	value := l.requiredString(key)
	if value == "" {
		return 0
	}

	hours, err := strconv.Atoi(value)
	if err != nil || hours <= 0 {
		l.errs = append(l.errs, fmt.Errorf("%s must be a positive integer", key))
		return 0
	}

	return time.Duration(hours) * time.Hour
}

func (l *envLoader) err() error {
	if len(l.errs) == 0 {
		return nil
	}

	return errors.Join(l.errs...)
}
