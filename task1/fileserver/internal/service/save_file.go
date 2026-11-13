package service

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
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

func (s FileSaver) listAllFileNames(folder string) (fileNames []string, err error) {
	folderEntries, err := os.ReadDir(path.Join(s.baseDir, folder))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []string{}, nil
		}
		return []string{}, fmt.Errorf("reading dir: %w", err)
	}

	var allFileNames []string

	for _, file := range folderEntries {
		if !file.IsDir() {
			allFileNames = append(allFileNames, file.Name())
		}
	}

	slices.SortFunc(allFileNames, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return allFileNames, nil
}

func (s *FileSaver) SaveFile(folder, name string, src io.Reader) (err error) {
	dir := filepath.Join(s.baseDir, folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create folder: %w", err)
	}

	pathToFile := filepath.Join(dir, name)
	dst, err := os.OpenFile(pathToFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)

	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	defer func() {
		if cerr := dst.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
		if err != nil {
			_ = os.Remove(pathToFile)
		}
	}()

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (s *FileSaver) ReplaceFile(folder, name string, src io.Reader) (err error) {
	pathToFile := filepath.Join(s.baseDir, folder, name)
	if _, err := os.Stat(pathToFile); err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(pathToFile), ".upload-*")

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

	return os.Rename(tmp.Name(), pathToFile)
}

func (s *FileSaver) GetFileNames(folder string) (fileNames []string, err error) {
	return s.listAllFileNames(folder)
}

func (s *FileSaver) GetFileContent(folder, name string) (io.ReadCloser, error) {
	filenames, err := s.listAllFileNames(folder)
	if err != nil {
		return nil, err
	}
	var targetFileName string
	for _, v := range filenames {
		if v == name {
			targetFileName = v
			break
		}
	}
	if targetFileName == "" {
		return nil, fmt.Errorf("file %q not found in folder %q: %w", name, folder, os.ErrNotExist)
	}
	pathToFile := path.Join(s.baseDir, folder, targetFileName)
	file, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	return file, nil
}
