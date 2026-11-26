package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type ServerConfig struct {
	Port     string `yaml:"port" env:"PORT"`
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL"`
	LogFile  string `yaml:"log_file" env:"LOG_FILE"`
}

func MustLoad(configPath string) ServerConfig {
	var cfg ServerConfig
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
