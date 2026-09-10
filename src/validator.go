package main

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	FileSizeLimit    = errors.New("размер файла не допустимый")
	FileTypeError    = errors.New("не подходящий формат файла")
	FileNameNotEmpty = errors.New("файл не должен быть пустым")
)

const maxFileSize = 50 * 1024 * 1024

func ValidateMedia(filename string, size int64) error {
	if strings.TrimSpace(filename) == "" {
		return FileNameNotEmpty
	}
	if size <= 0 || size > maxFileSize {
		return FileSizeLimit
	}
	errType := strings.ToLower(filepath.Ext(filename))
	switch errType {
	case ".mp4", ".mov", ".webm", ".png", ".jpg":
		return nil
	default:
		return FileTypeError
	}
}
