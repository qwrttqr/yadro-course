package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	WordsAddress string     `yaml:"words_address" env:"WORDS_ADDRESS"`
	HttpServer   HttpServer `yaml:"http_server"`
	LogLevel     string     `yaml:"log_level" env:"LOG_LEVEL"`
	LogFile      string     `yaml:"log_file" env:"LOG_FILE"`
}

type HttpServer struct {
	Address string `yaml:"address" env:"HTTP_SERVER_ADDRESS"`
	Timeout string `yaml:"timeout" env:"HTTP_SERVER_TIMEOUT"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("cannot read config %q: %s", configPath, err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("cannot read config %q: %s", configPath, err)
		}
	}

	return cfg
}
