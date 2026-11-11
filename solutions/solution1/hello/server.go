package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port string `yaml:"port" env:"HELLO_PORT" env-default:"8888"`
}

func Ping(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("pong\n"))
	if err != nil {
		slog.Error("could not write response", "error", err)
	}
}

func Hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		slog.Error("could not get name - empty value")
		http.Error(w, "empty name", http.StatusBadRequest)
		return
	}
	if _, err := fmt.Fprintf(w, "Hello, %s!\n", name); err != nil {
		slog.Error("could not reply", "error", err)
	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "configuration file")
	flag.Parse()

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(err)
	}

	// http handlers
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", Ping)
	mux.HandleFunc("GET /hello", Hello)

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
