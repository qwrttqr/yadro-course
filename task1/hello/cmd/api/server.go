package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ilyakaznacheev/cleanenv"
	"yadro.com/course/internal/config"
	"yadro.com/course/internal/handler"
)

func main() {
	var cfg config.ServerConfig
	if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		log.Fatalf("error while reading config: %v", err)
	}

	helloHandler := handler.NewHelloHandler()
	router := handler.NewRouter(helloHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), router)
	if err != nil {
		log.Fatalf("error while starting http server: %v", err)
	}
}
