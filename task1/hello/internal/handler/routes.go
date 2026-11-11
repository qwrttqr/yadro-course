package handler

import "net/http"

func NewRouter(helloHandler *HelloHandler) http.Handler {
	rootMux := http.NewServeMux()
	v1 := http.NewServeMux()
	helloHandler.RegisterRoutes(v1)
	rootMux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
	return rootMux
}
