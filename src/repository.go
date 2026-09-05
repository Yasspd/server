package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FileMetadata struct {
	ID        int
	FileName  string
	FilePath  string
	Size      int64
	MIMEType  string
	SHA256    string
	CreatedAt time.Time
}
type FileRepo struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) *FileRepo {
	return &FileRepo{pool: pool}
}
func (r *FileRepo) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id SERIAL PRIMARY KEY,
		file_name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		size BIGINT NOT NULL,
		mime_type TEXT NOT NULL,
		sha256 TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("ошибка создания файла в бд: %w", err)
	}
	return nil
}
func (r *FileRepo) Save(ctx context.Context, meta *FileMetadata) error {
	query := `
		INSERT INTO files (file_name, file_path, size, mime_type, sha256)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, created_at;
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		meta.FileName, // $1
		meta.FilePath, // $2
		meta.Size,     // $5
		meta.MIMEType, // $3
		meta.SHA256,   // $4
	).Scan(&meta.ID, &meta.CreatedAt)
	if err != nil {
		return fmt.Errorf("не получилось скопировать данные в бд: %w", err)
	}
	return nil
}
func (r *FileRepo) GetById(ctx context.Context, id int) (*FileMetadata, error) {
	query := `
		SELECT id, file_name, file_path, size, mime_type, sha256, created_at,
		FROM files,
		WHERE id = $1;
	`

	var meta FileMetadata
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&meta.ID,
		&meta.FileName,
		&meta.FilePath,
		&meta.Size,
		&meta.MIMEType,
		&meta.SHA256,
		&meta.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("файл с id %d не найден: %w", id, err)
	}

	return &meta, nil
}
