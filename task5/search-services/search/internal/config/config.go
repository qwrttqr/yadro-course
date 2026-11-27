package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	SearchAddress string `yaml:"search_address" env:"SEARCH_ADDRESS"`
	LogLevel      string `yaml:"log_level" env:"LOG_LEVEL"`
	LogFile       string `yaml:"log_file" env:"LOG_FILE"`
	DBAddress     string `yaml:"db_address" env:"DB_ADDRESS"`
	WordsAddress  string `yaml:"words_address" env:"WORDS_ADDRESS"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
