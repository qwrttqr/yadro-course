package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"yadro.com/course/api/adapters/rest"
	"yadro.com/course/api/adapters/search"
	"yadro.com/course/api/adapters/update"
	"yadro.com/course/api/adapters/words"
	"yadro.com/course/api/config"
	"yadro.com/course/api/core"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	logger := mustMakeLogger(cfg.LogLevel)

	logger.Info("starting server")
	logger.Debug("debug messages are enabled")

	updateClient, err := update.NewClient(cfg.UpdateAddress)
	if err != nil {
		logger.Error("cannot init update adapter", "error", err)
		os.Exit(1)
	}
	wordsClient, err := words.NewClient(cfg.WordsAddress)
	if err != nil {
		logger.Error("cannot init words adapter", "error", err)
		os.Exit(1)
	}
	searchClient, err := search.NewClient(cfg.SearchAddress)
	if err != nil {
		logger.Error("cannot init search adapter", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("POST /api/db/update", rest.NewUpdateHandler(logger, updateClient))
	mux.Handle("GET /api/db/stats", rest.NewUpdateStatsHandler(logger, updateClient))
	mux.Handle("GET /api/db/status", rest.NewUpdateStatusHandler(logger, updateClient))
	mux.Handle("DELETE /api/db", rest.NewDropHandler(logger, updateClient))
	mux.Handle("GET /api/search", rest.NewSearchHandler(logger, searchClient))
	mux.Handle("GET /api/ping", rest.NewPingHandler(logger, map[string]core.Pinger{"words": wordsClient, "update": updateClient, "search": searchClient}))

	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		logger.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Error("erroneous shutdown", "error", err)
		}
	}()

	logger.Info("Running HTTP server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server closed unexpectedly", "error", err)
			return
		}
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
