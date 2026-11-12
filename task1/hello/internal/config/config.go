package config

type ServerConfig struct {
	Port     string `yaml:"port" env:"PORT"`
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL"`
	LogFile  string `yaml:"log_file" env:"LOG_FILE"`
}
