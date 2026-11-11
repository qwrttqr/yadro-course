package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	fileStorePath = "/tmp/filestore"
	filePerm      = 0o664
	storePerm     = 0o775
	fileKey       = "file"
	nameKey       = "name"
)

type Config struct {
	Port string `yaml:"port" env:"FILESERVER_PORT" env-default:"8888"`
}

func init() {
	if err := os.MkdirAll(fileStorePath, storePerm); err != nil {
		panic(err)
	}
}

func closeOrLog(closer io.Closer) {
	if err := closer.Close(); err != nil {
		slog.Error("could not close", "error", err)
	}
}

func ListFiles(w http.ResponseWriter, r *http.Request) {
	slog.Info("list files")
	entries, err := os.ReadDir(fileStorePath)
	if err != nil {
		slog.Error("could not read file store", "error", err)
		http.Error(w, "cannot read file store", http.StatusInternalServerError)
		return
	}

	for _, entry := range entries {
		if _, err = fmt.Fprintln(w, entry.Name()); err != nil {
			slog.Error("could not write response", "error", err)
			return
		}
	}
}

func CreateFile(w http.ResponseWriter, r *http.Request) {
	slog.Info("create file")
	formReader, header, err := r.FormFile(fileKey)
	if err != nil {
		slog.Error("could not read file form", "error", err)
		http.Error(w, "could not read file form", http.StatusBadRequest)
		return
	}
	defer closeOrLog(formReader)

	root, err := os.OpenRoot(fileStorePath)
	if err != nil {
		slog.Error("could not read file store", "error", err)
		http.Error(w, "cannot read file store", http.StatusInternalServerError)
		return
	}
	defer closeOrLog(root)

	filePath := header.Filename

	if _, err = root.Stat(filePath); !os.IsNotExist(err) {
		slog.Error("file must be absent to be created", "filepath", filePath, "error", err)
		http.Error(w, "file already exists", http.StatusConflict)
		return
	}

	fileWriter, err := root.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, filePerm)
	if err != nil {
		slog.Error("cannot create file", "filepath", filePath, "error", err)
		http.Error(w, "cannot create file", http.StatusInternalServerError)
		return
	}
	defer closeOrLog(fileWriter)

	_, err = io.Copy(fileWriter, formReader)
	if err != nil {
		slog.Error("cannot write file", "filepath", filePath, "error", err)
		http.Error(w, "cannot write file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if _, err = fmt.Fprintln(w, filePath); err != nil {
		slog.Error("could not write response", "error", err)
	}
}

func ReadFile(w http.ResponseWriter, r *http.Request) {
	slog.Info("read file")
	filePath := r.PathValue(nameKey)
	root, err := os.OpenRoot(fileStorePath)
	if err != nil {
		slog.Error("could not read file store", "error", err)
		http.Error(w, "cannot read file store", http.StatusInternalServerError)
		return
	}
	defer closeOrLog(root)

	stat, err := root.Stat(filePath)
	os.IsNotExist(err)
	if err != nil {
		slog.Error("file not found")
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-length", strconv.FormatInt(stat.Size(), 10))

	fileReader, err := root.Open(filePath)
	if err != nil {
		slog.Error("cannot open file", "filepath", filePath, "error", err)
		http.Error(w, "cannot open file", http.StatusInternalServerError)
		return
	}
	defer closeOrLog(fileReader)
	_, err = io.Copy(w, fileReader)
	if err != nil {
		slog.Error("cannot read file", "filepath", filePath, "error", err)
	}
}

func UpdateFile(w http.ResponseWriter, r *http.Request) {
	slog.Info("update file")
	formReader, _, err := r.FormFile(fileKey)
	if err != nil {
		slog.Error("could not read file form", "error", err)
		http.Error(w, "could not read file form", http.StatusBadRequest)
		return
	}
	defer closeOrLog(formReader)

	filePath := r.PathValue(nameKey)
	root, err := os.OpenRoot(fileStorePath)
	if err != nil {
		slog.Error("could not read file store", "error", err)
		http.Error(w, "cannot read file store", http.StatusInternalServerError)
		return
	}
	defer closeOrLog(root)

	if _, err = root.Stat(filePath); os.IsNotExist(err) {
		slog.Error("file must be available to be rewritten")
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	fileWriter, err := root.OpenFile(filePath, os.O_TRUNC|os.O_WRONLY, filePerm)
	if err != nil {
		slog.Error("cannot create file", "filepath", filePath, "error", err)
		http.Error(w, "cannot create file", http.StatusInternalServerError)
		return
	}
	defer closeOrLog(fileWriter)

	_, err = io.Copy(fileWriter, formReader)
	if err != nil {
		slog.Error("cannot write file", "filepath", filePath, "error", err)
		http.Error(w, "cannot write file", http.StatusInternalServerError)
		return
	}
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	slog.Info("delete file")
	filePath := r.PathValue(nameKey)

	root, err := os.OpenRoot(fileStorePath)
	if err != nil {
		slog.Error("could not read file store", "error", err)
		http.Error(w, "cannot read file store", http.StatusInternalServerError)
		return
	}
	defer closeOrLog(root)

	if _, err = root.Stat(filePath); os.IsNotExist(err) {
		slog.Info("file to delete is not found", "filepath", filePath)
		return
	}

	if err := root.Remove(filePath); err != nil {
		slog.Error("cannot remove file", "filepath", filePath, "error", err)
		http.Error(w, "cannot remove file", http.StatusInternalServerError)
	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "configuration file")
	flag.Parse()

	var cfg Config
	if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		panic(err)
	}

	// http handlers
	mux := http.NewServeMux()
	mux.HandleFunc("GET /files", ListFiles)
	mux.HandleFunc("POST /files", CreateFile)
	mux.HandleFunc("GET /files/{name}", ReadFile)
	mux.HandleFunc("PUT /files/{name}", UpdateFile)
	mux.HandleFunc("DELETE /files/{name}", DeleteFile)

	// http server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("http server failed", "error", err)
	}
}
