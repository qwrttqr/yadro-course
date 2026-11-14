package service

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FileSaver struct {
	baseDir string
}

func NewFileSaver(baseDir string) *FileSaver {
	return &FileSaver{baseDir: baseDir}
}

func (s *FileSaver) listAllFileNames() (fileNames []string, err error) {
	folderEntries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []string{}, nil
		}
		return []string{}, fmt.Errorf("reading dir: %w", err)
	}

	allFileNames := make([]string, 0, len(folderEntries))

	for _, file := range folderEntries {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			allFileNames = append(allFileNames, file.Name())
		}
	}

	return allFileNames, nil
}

func (s *FileSaver) SaveFile(name string, src io.Reader) (err error) {
	pathToFile := filepath.Join(s.baseDir, name)
	tmp, err := os.CreateTemp(filepath.Dir(pathToFile), ".upload-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}

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

func (s *FileSaver) ReplaceFile(name string, src io.Reader) (err error) {
	pathToFile := filepath.Join(s.baseDir, name)
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

func (s *FileSaver) GetFileNames() (fileNames []string, err error) {
	return s.listAllFileNames()
}

func (s *FileSaver) GetFileContent(name string) (io.ReadCloser, error) {
	filenames, err := s.listAllFileNames()
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
		return nil, fmt.Errorf("file %q not found: %w", name, os.ErrNotExist)
	}
	pathToFile := filepath.Join(s.baseDir, targetFileName)
	file, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *FileSaver) DeleteFile(name string) error {
	pathToFile := filepath.Join(s.baseDir, name)
	err := os.Remove(pathToFile)
	return err
}
