package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HelloHandler struct {
}

func NewHelloHandler() *HelloHandler {
	return &HelloHandler{}
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
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *HelloHandler) helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hello %s", name)))
}
