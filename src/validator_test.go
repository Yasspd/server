package main

import (
	"errors"
	"testing"
)

func TestValidateMedia(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		size        int64
		expectedErr error
	}{

		{
			name:        "Корректный MP4 файл",
			filename:    "promo.mp4",
			size:        1024 * 1024, // 1 MB
			expectedErr: nil,
		},
		{
			name:        "Файл с пустым именем",
			filename:    "   ",
			size:        1024,
			expectedErr: FileNameNotEmpty,
		},
		{
			name:        "Недопустимое расширение (.exe)",
			filename:    "virus.exe",
			size:        1024,
			expectedErr: FileTypeError,
		},
		{
			name:        "Превышение максимального размера",
			filename:    "huge_movie.mp4",
			size:        60 * 1024 * 1024, // 60 MB
			expectedErr: FileSizeLimit,
		},
		{
			name:        "Файл с нулевым размером",
			filename:    "empty.jpg",
			size:        0,
			expectedErr: FileSizeLimit,
		},
	}
	for _, i := range tests {
		t.Run(i.name, func(t *testing.T) {
			err := ValidateMedia(i.filename, i.size)
			if !errors.Is(err, i.expectedErr) {
				t.Errorf("ValidateMedia() = ошибка %v, ожидали %v", err, i.expectedErr)
			}

		})
	}
}
