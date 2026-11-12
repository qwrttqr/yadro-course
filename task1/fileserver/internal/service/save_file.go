package service

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type FileSaver struct {
	baseDir string
}

func NewFileSaver(baseDir string) *FileSaver {
	return &FileSaver{baseDir: baseDir}
}

func (s *FileSaver) SaveFile(folder, name string, src io.Reader) (err error) {
	dir := filepath.Join(s.baseDir, folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create folder: %w", err)
	}

	path := filepath.Join(dir, name)
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)

	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	defer func() {
		if cerr := dst.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
		if err != nil {
			_ = os.Remove(path)
		}
	}()

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (s *FileSaver) ReplaceFile(folder, name string, src io.Reader) (err error) {
	path := filepath.Join(s.baseDir, folder, name)
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".upload-*")

	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}

	defer func() {
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()

	if _, err = io.Copy(tmp, src); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}

	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temp: %w", err)
	}

	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	if err = os.Chmod(tmp.Name(), 0o644); err != nil {
		return fmt.Errorf("chmod temp: %w", err)
	}

	return os.Rename(tmp.Name(), path)
}

func (s *FileSaver) GetFileNames() (fileNames []string, err error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []string{}, nil
		}
		return []string{}, fmt.Errorf("reading dir: %w", err)
	}
	var allFileNames []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDirPath := filepath.Join(s.baseDir, entry.Name())
		files, err := os.ReadDir(subDirPath)

		if err != nil {
			return nil, fmt.Errorf("reading subdir %s: %w", subDirPath, err)
		}

		for _, file := range files {
			if !file.IsDir() {
				allFileNames = append(allFileNames, file.Name())
			}
		}
	}

	slices.SortFunc(allFileNames, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return allFileNames, nil
}
