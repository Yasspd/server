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

func (s *StorageService) SaveUploadedFile(
	ctx context.Context,
	fileName string,
	mimeType string,
	src io.Reader,
) (*FileMetadata, error) {

	// 1. Формируем уникальное имя и полный путь на диске
	safeName := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(fileName))
	fullPath := filepath.Join(s.uploadDir, safeName)

	// 2. Создаем файл на диске
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать файл на диске: %w", err)
	}
	defer dst.Close()

	// 3. Создаем хэшер и тройник TeeReader
	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)

	// 4. Потоковая запись: читает из tee -> считает хэш -> пишет в dst
	size, err := io.Copy(dst, tee)
	if err != nil {
		return nil, fmt.Errorf("ошибка записи потока данных: %w", err)
	}

	// 5. Фиксируем SHA-256 хэш
	hashSum := hex.EncodeToString(hasher.Sum(nil))

	// 6. Формируем объект метаданных и сохраняем в базу данных
	meta := &FileMetadata{
		FileName: fileName,
		FilePath: fullPath,
		Size:     size,
		MIMEType: mimeType,
		SHA256:   hashSum,
	}

	if err := s.repo.Save(ctx, meta); err != nil {
		// Если упала база — подчищаем за собой созданный файл с диска
		_ = os.Remove(fullPath)
		return nil, fmt.Errorf("не удалось сохранить метаданные: %w", err)
	}

	return meta, nil
}
