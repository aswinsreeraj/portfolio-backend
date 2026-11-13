package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

// Config holds Postgres connection options
type Config struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ConnectPostgres establishes a database connection pool with health check
func ConnectPostgres(cfg Config) (*sql.DB, error) {
	connStr := cfg.URL
	if connStr == "" {
		connStr = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=5",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	if cfg.URL != "" {
		slog.Info("PostgreSQL connection established successfully via DATABASE_URL")
	} else {
		slog.Info("PostgreSQL connection established successfully", "host", cfg.Host, "port", cfg.Port, "db", cfg.DBName)
	}
	return db, nil
}

