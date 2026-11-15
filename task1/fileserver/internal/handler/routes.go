package handler

import "net/http"

func NewRouter(handler *FileHandler) http.Handler {
	rootMux := http.NewServeMux()
	handler.RegisterRoutes(rootMux)
	//handler.RegisterRoutes(v1)
	//rootMux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
	return rootMux
}
