package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)
	fmt.Println(cfg)

	log := mustMakeLogger("log level from config")

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	wordsClient, err := words.NewClient("address of words service", log)
	if err != nil {
		log.Error("cannot init words adapter", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	// mux.Handle("GET /api/words", rest.NewWordsHandler(...)) to be implemented
	mux.Handle("GET /ping", rest.NewPingHandler(log, map[string]core.Pinger{"words": wordsClient}))

	server := http.Server{
		Addr:        "localhost:8888",    // replace with address from config
		ReadTimeout: 10000 * time.Second, // replace with timeout from config
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}

}

func mustMakeLogger(logLevel string) *slog.Logger {
	return slog.Default()
}
