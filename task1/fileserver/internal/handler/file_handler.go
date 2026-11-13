package handler

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
)

const maxUploadSize = 20 << 20

var (
	validFolderRegex   = regexp.MustCompile(`^[a-zA-Z]+$`)
	validFileNameRegex = regexp.MustCompile(`^[\p{L}\p{N}_-][\p{L}\p{N}._ -]*$`)
)

type fileSaver interface {
	SaveFile(folder, name string, src io.Reader) error
	ReplaceFile(folder, name string, src io.Reader) error
	GetFileNames(folder string) (fileNames []string, err error)
	GetFileContent(folder, name string) (io.ReadCloser, error)
}

type FileHandler struct {
	logger *slog.Logger
	saver  fileSaver
}

func NewFileHandler(logger *slog.Logger, saver fileSaver) *FileHandler {
	return &FileHandler{logger: logger, saver: saver}
}

func (h *FileHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /files", h.uploadNew)
	mux.HandleFunc("POST /files/{filename}", h.updateExisting)
	mux.HandleFunc("PUT /files/{filename}", h.updateExisting)
	mux.HandleFunc("GET /files", h.getFiles)
	mux.HandleFunc("GET /files/{filename}", h.getFileContent)
}

type uploadResponse struct {
	Name string `json:"name"`
}

func (h *FileHandler) uploadNew(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			http.Error(w, "file too big", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	folder := r.FormValue("folder")
	if !validFolderRegex.MatchString(folder) {
		http.Error(w, "invalid folder name", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing or invalid file", http.StatusBadRequest)
		return
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			h.logger.Error("close uploaded file", "error", cerr)
		}
	}()

	filename := header.Filename
	if len(filename) > 255 || !validFileNameRegex.MatchString(filename) {
		http.Error(w, "invalid file name", http.StatusBadRequest)
		return
	}

	if err := h.saver.SaveFile(folder, filename, file); err != nil {
		if errors.Is(err, fs.ErrExist) {
			http.Error(w, "file already exists", http.StatusConflict)
			return
		}
		h.logger.Error("save file", "error", err, "folder", folder, "name", filename)
		http.Error(w, "error saving file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(uploadResponse{Name: filename}); err != nil {
		h.logger.Error("encode response", "error", err)
	}
}

func (h *FileHandler) updateExisting(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			http.Error(w, "file too big", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	filename := r.PathValue("filename")
	folder := r.FormValue("folder")
	if !validFolderRegex.MatchString(folder) {
		http.Error(w, "invalid folder name", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("missing or invalid file", "error", err, "folder", folder, "name", filename)
		http.Error(w, "missing or invalid file", http.StatusBadRequest)
		return
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			h.logger.Error("close uploaded file", "error", cerr)
		}
	}()

	if len(filename) > 255 || !validFileNameRegex.MatchString(filename) {
		http.Error(w, "invalid file name", http.StatusBadRequest)
		return
	}

	if err := h.saver.ReplaceFile(folder, filename, file); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		h.logger.Error("update file", "error", err, "folder", folder, "name", filename)
		http.Error(w, "error update file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

type getFilesResponse struct {
	FileNames []string `json:"data"`
}

func (h *FileHandler) getFiles(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")

	if !validFolderRegex.MatchString(folder) {
		http.Error(w, "invalid folder name", http.StatusBadRequest)
		return
	}

	filenames, err := h.saver.GetFileNames(folder)
	if err != nil {
		h.logger.Error("read files", "error", err)
		http.Error(w, "error reading files", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(getFilesResponse{FileNames: filenames}); err != nil {
		h.logger.Error("encode response", "error", err)
	}
}

func (h *FileHandler) getFileContent(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")

	if !validFolderRegex.MatchString(folder) {
		http.Error(w, "invalid folder name", http.StatusBadRequest)
		return
	}

	filename := r.PathValue("filename")

	if len(filename) > 255 || !validFileNameRegex.MatchString(filename) {
		http.Error(w, "invalid file name", http.StatusBadRequest)
		return
	}

	file, err := h.saver.GetFileContent(folder, filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		h.logger.Error("read files", "error", err)
		http.Error(w, "error reading files", http.StatusInternalServerError)
		return
	}

	if file != nil {
		defer func() {
			if cerr := file.Close(); cerr != nil {
				h.logger.Error("error file close", "error", err)
			}
		}()
	}

	contentType := mime.TypeByExtension(filepath.Ext(filename))

	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)

	if _, err := io.Copy(w, file); err != nil {
		h.logger.Error("stream file", "error", err)
	}
}
