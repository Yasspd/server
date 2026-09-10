package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const uploadDir = "./uploading"

type FileHandler struct {
	storage *StorageService
}

func NewFileHandler(storage *StorageService) *FileHandler {
	return &FileHandler{storage: storage}
}

func enableCors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func (h *FileHandler) Streamfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	idstr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, " не получилось найти файл по id", http.StatusNotFound)
		return
	}
	file, meta, err := h.storage.GetFileForStream(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	if meta.MIMEType != "" {
		w.Header().Set("Content-type", meta.MIMEType)
	}

	http.ServeContent(w, r, meta.FileName, meta.CreatedAt, file)
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "не тот формат запроса", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(300 << 20); err != nil {
		http.Error(w, "файл больше 10мб"+err.Error(), http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "не получилось получить файл"+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	meta, err := h.storage.SaveUploadedFile(
		r.Context(),
		handler.Filename,
		handler.Header.Get("Context-type"),
		file,
	)
	if err != nil {
		http.Error(w, "Не получилось сохранить файл в хранилище"+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Context-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"status":"ok", "id":%d, "sha256":"%s"}`, meta.ID, meta.SHA256)))
}

func main() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("Не удалось создать папку загрузок: %v", err)
	}
	dsn := "postgres://petya_db:arelun06_db@localhost:5666/Data_cloud"
	pool, err := ConnectionDB(dsn)
	if err != nil {
		log.Fatalf("не удалось подключиться к базе данных: %v", err)
		return
	}
	repo := NewFileRepository(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.InitSchema(ctx); err != nil {
		log.Fatalf("Критическая ошибка: не удалось создать таблицу files: %v", err)
	}
	log.Println("Таблица files успешно проверена/создана в БД")

	repo.InitSchema(context.Background())
	storage := NewStorageService(repo, uploadDir)
	handler := NewFileHandler(storage)
	http.HandleFunc("/api/v1/upload", enableCors(handler.Upload))
	http.HandleFunc("/api/v1/stream", enableCors(handler.Streamfile))
	fmt.Println("Storage Service запущен на порту :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
