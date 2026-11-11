package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	petnamepb "yadro.com/course/proto"
)

const (
	// usually streams are not designed to have timeout ;)
	streamServingTimeout = 10 * time.Second
	// max wait for shutdown before forcing it
	maxShutdownTime = 5 * time.Second
)

type Config struct {
	Port string `json:"port"`
}

type server struct {
	petnamepb.UnimplementedPetnameGeneratorServer
	log *slog.Logger
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *server) Generate(
	_ context.Context, in *petnamepb.PetnameRequest,
) (*petnamepb.PetnameResponse, error) {
	words := int(in.GetWords())
	separator := in.GetSeparator()
	s.log.Debug("generate", "words", words, "separator", separator)
	if words <= 0 {
		return nil, status.Error(codes.InvalidArgument, "number of words must be bigger than 0")
	}
	name := petname.Generate(words, separator)
	return &petnamepb.PetnameResponse{Name: name}, nil
}

func (s *server) GenerateMany(
	in *petnamepb.PetnameStreamRequest,
	stream petnamepb.PetnameGenerator_GenerateManyServer,
) error {
	words := int(in.GetWords())
	names := in.GetNames()
	separator := in.GetSeparator()
	s.log.Debug("generate many", "words", words, "separator", separator, "names", names)
	if words <= 0 {
		s.log.Error("wrong number of words", "words", words)
		return status.Error(codes.InvalidArgument, "number of words must be bigger than 0")
	}
	if names <= 0 {
		s.log.Error("wrong number of names", "names", words)
		return status.Error(codes.InvalidArgument, "number of names must be bigger than 0")
	}

	ctx, cancel := context.WithTimeout(stream.Context(), streamServingTimeout)
	defer cancel()

	for range names {
		if ctx.Err() != nil {
			s.log.Error("context done while serving", "error", ctx.Err)
			return ctx.Err()
		}
		if err := stream.Send(&petnamepb.PetnameResponse{
			Name: petname.Generate(words, separator),
		}); err != nil {
			s.log.Error("could not send name", "error", err)
			return err
		}
	}
	return nil
}

func main() {

	// in main() we only read config and setup logger

	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "configuration file")
	flag.Parse()

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(err)
	}
	log := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
		),
	)

	if err := run(cfg, log); err != nil {
		log.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}

func run(cfg Config, log *slog.Logger) error {
	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen port %s: %v", cfg.Port, err)
	}

	s := grpc.NewServer()
	petnamepb.RegisterPetnameGeneratorServer(s, &server{log: log})
	reflection.Register(s)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		log.Info("starting server", "port", cfg.Port)
		if err = s.Serve(listener); err != nil {
			log.Error("faied to serve", "error", err)
			// notify shutdown code if needed
			cancel()
		}
	}()

	// got failed serve or terminate signal
	<-ctx.Done()

	// force after some time
	time.AfterFunc(maxShutdownTime, func() {
		log.Info("forcing server stop")
		s.Stop()
	})

	log.Info("starting graceful stop")
	s.GracefulStop()
	log.Info("server stopped")

	return err
}
