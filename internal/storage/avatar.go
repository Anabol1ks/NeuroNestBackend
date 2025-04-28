package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type AvatarService interface {
	Save(reader io.Reader, filename string) (publicURL string, err error)
	Delete(filename string) error
}

type LocalAvatarService struct {
	BasePath string // "./uploads/avatars"
	BaseURL  string // "http://localhost:8080/avatars"
}

func NewLocalAvatarService(basePath, baseURL string) *LocalAvatarService {
	return &LocalAvatarService{BasePath: basePath, BaseURL: baseURL}
}

func (s *LocalAvatarService) Save(reader io.Reader, filename string) (string, error) {
	// создаём каталог, если не существует
	if err := os.MkdirAll(s.BasePath, 0o755); err != nil {
		return "", err
	}

	dstPath := filepath.Join(s.BasePath, filename)
	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return "", err
	}

	// возвращаем полный публичный URL
	return fmt.Sprintf("%s/%s", strings.TrimRight(s.BaseURL, "/"), filename), nil
}

func (s *LocalAvatarService) Delete(filename string) error {
	return os.Remove(filepath.Join(s.BasePath, filename))
}
