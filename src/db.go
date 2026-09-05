package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectionDB(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга DSN: %w", err)
	}
	cfg.MaxConns = 20
	cfg.MinConns = 5
	cfg.MaxConnIdleTime = 20 * time.Minute
	cfg.MaxConnLifetime = 15 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("база данных недоступна: %w", err)
	}

	return pool, nil
}
