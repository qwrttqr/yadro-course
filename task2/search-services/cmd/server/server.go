package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"yadro.com/course/internal/config"
	"yadro.com/course/internal/grpcserver"
	"yadro.com/course/internal/logging"
	"yadro.com/course/internal/service"

	wordspb "yadro.com/course/proto/words"
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

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))

	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	defer func() {
		if cerr := listener.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("error stopping listener %v", cerr))
		}
	}()

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

	wordsNormalizationService := service.NewWordsNormalizationService()

	if err = cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		return fmt.Errorf("error while reading config: %v", err)
	}

	server := grpcserver.NewWordsServer(wordsNormalizationService, logger)
	s := grpc.NewServer(grpc.MaxRecvMsgSize(8 << 10))
	wordspb.RegisterWordsServer(s, server)
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("error while start serving: %v", err)
	}
	return err
}
