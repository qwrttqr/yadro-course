package handler

import (
	"encoding/json"
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

type helloHandlerResponse struct {
	Message string `json:"message"`
}

func (h *HelloHandler) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := helloHandlerResponse{Message: "pong"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("error while executing request", "err", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *HelloHandler) helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "Hello %s", name)
	if err != nil {
		h.logger.Error("error while executing request", "err", err)
	}
}
