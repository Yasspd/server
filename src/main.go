package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const uploadDir = "../uploading"

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

func HandlerUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	}

	err := r.ParseMultipartForm(10 << 10)
	if err != nil {
		http.Error(w, "Максимальный обьем файла не должно превышать больше 10мб"+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "не верный тип файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("получен файл, размер: %d байт", handler.Filename, handler.Size)

	safeFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(handler.Filename))
	dstPath := filepath.Join(uploadDir, safeFileName)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "не получилось создать файл на сервере"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	Writenbytes, err := io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Ошибка записи потока на диск: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Успешно сохранено: %s (%d байт)", dstPath, Writenbytes)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"status":"ok", "filename":"%s"}`, safeFileName)))
}

func main() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("Не удалось создать папку загрузок: %v", err)
	}
	http.HandleFunc("/api/v1/upload", enableCors(HandlerUpload))
	fmt.Println("Storage Service запущен на порту :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
