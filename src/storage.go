package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type StorageService struct {
	repo      *FileRepo
	uploadDir string
}

func NewStorageService(repo *FileRepo, uploadDir string) *StorageService {
	return &StorageService{
		repo:      repo,
		uploadDir: uploadDir,
	}
}

func (s *StorageService) GetFileForStream(ctx context.Context, id int) (*os.File, *FileMetadata, error) {
	meta, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("Не найден файл: %w", err)

	}
	readFile, err := os.Open(meta.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Файл не найден на диске или не удалось его прочитать: %w", err)
	}
	return readFile, meta, nil
}

func (s *StorageService) SaveUploadedFile(
	ctx context.Context,
	fileName string,
	mimeType string,
	src io.Reader,
) (*FileMetadata, error) {
	safeName := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(fileName))
	fullPath := filepath.Join(s.uploadDir, safeName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать файл на диске: %w", err)
	}
	defer dst.Close()

	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)

	size, err := io.Copy(dst, tee)
	if err != nil {
		return nil, fmt.Errorf("ошибка записи потока данных: %w", err)
	}

	hashSum := hex.EncodeToString(hasher.Sum(nil))

	meta := &FileMetadata{
		FileName: fileName,
		FilePath: fullPath,
		Size:     size,
		MIMEType: mimeType,
		SHA256:   hashSum,
	}

	if err := s.repo.Save(ctx, meta); err != nil {
		_ = os.Remove(fullPath)
		return nil, fmt.Errorf("не удалось сохранить метаданные: %w", err)
	}

	return meta, nil
}
