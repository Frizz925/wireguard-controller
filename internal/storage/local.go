package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"
)

const DEFAULT_STORAGE_DIR = "data"

type LocalStorage struct {
	Directory string
}

func NewLocalStorage(dirs ...string) *LocalStorage {
	dir := DEFAULT_STORAGE_DIR
	if len(dirs) > 0 {
		dir = path.Join(dirs...)
	}
	return &LocalStorage{
		Directory: dir,
	}
}

func (s *LocalStorage) List(ctx context.Context, prefix ...string) ([]string, error) {
	dir := s.Directory
	if len(prefix) > 0 {
		dir = path.Join(s.Directory, prefix[0])
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	results := make([]string, 0)
	for _, file := range files {
		name := file.Name()
		idx := strings.Index(name, ".json")
		if idx <= 0 {
			continue
		}
		results = append(results, name[:idx])
	}
	return results, nil
}

func (s *LocalStorage) Load(ctx context.Context, name string, v any) error {
	f, err := os.Open(s.getFilePath(name))
	if err != nil {
		return err
	}
	defer f.Close()
	return s.unmarshal(f, v)
}

func (s *LocalStorage) Save(ctx context.Context, name string, data any) error {
	filePath := s.getFilePath(name)
	if err := s.ensureDirectory(path.Dir(filePath)); err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(data)
}

func (s *LocalStorage) Delete(ctx context.Context, name string) error {
	return os.Remove(s.getFilePath(name))
}

func (s *LocalStorage) ensureDirectory(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(dir, 0700)
}

func (s *LocalStorage) getFilePath(name string) string {
	return path.Join(s.Directory, name+".json")
}

func (s *LocalStorage) unmarshal(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}
