package config

type ServerConfig struct {
	Port string `yaml:"port" env:"FILESERVER_PORT"`
}
