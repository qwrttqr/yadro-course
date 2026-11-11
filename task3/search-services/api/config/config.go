package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
