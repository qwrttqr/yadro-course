package config

type ServerConfig struct {
	WordsAddress string `yaml:"words_address" env:"WORDS_ADDRESS"`
	LogLevel     string `yaml:"log_level" env:"LOG_LEVEL"`
	LogFile      string `yaml:"log_file" env:"LOG_FILE"`
}
