package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"yadro.com/course/internal/config"
	"yadro.com/course/internal/handler"
	"yadro.com/course/internal/logging"
	"yadro.com/course/internal/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (err error) {
	var cfg config.ServerConfig
	var logger *slog.Logger
	var f *os.File

	if err = cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		return fmt.Errorf("error while reading config: %v", err)
	}

	f, err = os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

	if err != nil {
		return fmt.Errorf("error opening log file %v", err)
	}

	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("error closing log file %v", cerr))
		}
	}()

	logger, err = logging.New(cfg.LogLevel, f)

	if err != nil {
		return fmt.Errorf("error creating logger %v", err)
	}

	fileHandler := handler.NewFileHandler(logger, service.NewFileSaver(cfg.BaseDir))
	router := handler.NewRouter(fileHandler)
	err = http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), router)

	if err != nil {
		return fmt.Errorf("error while starting http server: %v", err)
	}
	return err
}
