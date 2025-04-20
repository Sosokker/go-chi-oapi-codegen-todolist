package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Sosokker/todolist-backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnectionPool(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Configure pool settings
	// poolConfig.MaxConns = 10
	// poolConfig.MinConns = 2
	// poolConfig.MaxConnIdleTime = 5 * time.Minute
	// poolConfig.MaxConnLifetime = 1 * time.Hour
	// poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	slog.Info("Database connection pool established")
	return pool, nil
}
