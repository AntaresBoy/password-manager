package store

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (f *FileStore) Read() ([]byte, error) {
	return os.ReadFile(f.path)
}

func (f *FileStore) Write(data []byte) error {
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	tmpPath := f.path + ".tmp"
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, f.path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

func (f *FileStore) Exists() bool {
	_, err := os.Stat(f.path)
	return !os.IsNotExist(err)
}

func (f *FileStore) Path() string {
	return f.path
}
