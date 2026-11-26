package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"yadro.com/course/api/adapters/rest"
	"yadro.com/course/api/adapters/words"
	"yadro.com/course/api/config"
	"yadro.com/course/api/core"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (err error) {
	var configPath string
	var log *slog.Logger
	var f *os.File
	flag.StringVar(&configPath, "config", "", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)
	fmt.Println(cfg)
	if cfg.LogFile != "" {
		f, err = os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				err = errors.Join(err, fmt.Errorf("error closing log file %v", cerr))
			}
		}()
		log, err = mustMakeLogger(cfg.LogLevel, f)

		if err != nil {
			return err
		}
	} else {
		log, err = mustMakeLogger(cfg.LogLevel, nil)

		if err != nil {
			return err
		}
	}

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	wordsClient, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("GET /api/words", rest.NewWordsHandler(log, wordsClient))
	mux.Handle("GET /api/ping", rest.NewPingHandler(log, map[string]core.Pinger{"words": wordsClient}))

	readTimeout, err := time.ParseDuration(cfg.HttpServer.Timeout)
	if err != nil {
		return fmt.Errorf("invalid HTTP server timeout %q: %w", cfg.HttpServer.Timeout, err)
	}

	server := http.Server{
		Addr:        cfg.HttpServer.Address,
		ReadTimeout: readTimeout,
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		err = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	return err
}

func mustMakeLogger(levelStr string, file io.Writer) (*slog.Logger, error) {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(levelStr)); err != nil {
		return nil, err
	}
	var handler slog.Handler
	if file != nil {
		handler = slog.NewJSONHandler(file, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	}
	return slog.New(handler), nil
}
