package handler

import (
	"fmt"
	"log/slog"
	"net/http"
)

type HelloHandler struct {
	logger *slog.Logger
}

func NewHelloHandler(logger *slog.Logger) *HelloHandler {
	return &HelloHandler{logger: logger}
}

func (h *HelloHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ping", h.pingHandler)
	mux.HandleFunc("GET /hello", h.helloHandler)
}

func (h *HelloHandler) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(w, "pong"); err != nil {
		h.logger.Error("write response", "error", err)
		return
	}
}

func (h *HelloHandler) helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "empty name", http.StatusBadRequest)
		return
	}
	_, err := fmt.Fprintf(w, "Hello, %s!\n", name)
	if err != nil {
		h.logger.Error("error while executing request", "err", err)
	}
}
