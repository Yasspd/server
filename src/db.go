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

	// 1. Парсим DSN строку в объект конфигурации
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга DSN: %w", err)
	}

	// 2. Задаем параметры пула соединений (типизация через time.Duration)
	cfg.MaxConns = 20
	cfg.MinConns = 5
	cfg.MaxConnIdleTime = 20 * time.Minute
	cfg.MaxConnLifetime = 15 * time.Minute

	// 3. Создаем пул с нашей конфигурацией
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула: %w", err)
	}

	// 4. Проверяем доступность базы
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("база данных недоступна: %w", err)
	}

	return pool, nil
}
