package filestore

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type FileStore interface {
	Upload(file multipart.File, objectName string) (string, error)
	GetURL(objectName string) string
}

type LocalFileStore struct {
	basePath string // e.g., "./static"
	baseURL  string // e.g., "/static"
}

func NewLocalFileStore(basePath, baseURL string) *LocalFileStore {
	return &LocalFileStore{basePath: basePath, baseURL: baseURL}
}

func (l *LocalFileStore) Upload(file multipart.File, objectName string) (string, error) {
	fullPath := filepath.Join(l.basePath, objectName)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return l.GetURL(objectName), nil
}

func (l *LocalFileStore) GetURL(objectName string) string {
	// Use forward slashes for URLs
	return filepath.ToSlash(filepath.Join(l.baseURL, objectName))
}
