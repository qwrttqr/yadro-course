package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/internal/adapters/db"
	grpc2 "yadro.com/course/search/internal/adapters/grpc"
	"yadro.com/course/search/internal/config"
	"yadro.com/course/search/internal/service"
)

func main() {
	// config
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("error closing log file %v", cerr))
		}
	}()
	if err != nil {
		log.Fatalf("error opening log file %v", err)
	}

	// logger
	logger := mustMakeLogger(cfg.LogLevel, f)

	if err = run(cfg, logger); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, logger *slog.Logger) (err error) {

	listener, err := net.Listen("tcp", cfg.SearchAddress)

	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	defer func() {
		if cerr := listener.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("error stopping listener %v", cerr))
		}
	}()

	dbClient, err := db.New(logger, cfg.DBAddress)
	if err != nil {
		return err
	}
	searchService := service.NewSearchService(dbClient, logger)

	server := grpc2.NewSearchServer(searchService, logger)
	s := grpc.NewServer(grpc.MaxRecvMsgSize(8 << 10))
	searchpb.RegisterSearchServer(s, server)
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("error while start serving: %v", err)
	}
	return err
}

func mustMakeLogger(logLevel string, f io.Writer) *slog.Logger {
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
	var handler slog.Handler
	if f != nil {
		handler = slog.NewJSONHandler(f, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}
	return slog.New(handler)
}
