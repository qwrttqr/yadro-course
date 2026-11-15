package handler

import "net/http"

func NewRouter(handler *HelloHandler) http.Handler {
	rootMux := http.NewServeMux()
	handler.RegisterRoutes(rootMux)
	return rootMux
}
