package config

import (
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerTimeout int    `yaml:"server_shutdown_timeout"`
	DbTimeout     int    `yaml:"db_connect_timeout"`
	Port          int    `yaml:"port"`
	LoggerLevel   string `yaml:"logger_level"`
	DatabaseURL   string
}

func New() (Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	_ = godotenv.Load()

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "postgres://kinosearch:kinosearch@localhost:5432/kinosearch?sslmode=disable"
	}

	return cfg, nil
}
