package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	DBHost     string `env:"DB_HOST" env-default:"localhost"`
	DBPort     string `env:"DB_PORT" env-default:"5432"`
	DBUser     string `env:"DB_USER" env-default:"postgres"`
	DBPassword string `env:"DB_PASSWORD" env-default:"postgres"`
	DBName     string `env:"DB_NAME" env-default:"subscriptions"`
	DBSSLMode  string `env:"DB_SSLMODE" env-default:"disable"`
	HTTPPort   string `env:"HTTP_PORT" env-default:"8080"`
}

const defaultConfig = ".env"

func LoadConfig() *Config {

	// 1. Берём путь из ENV, иначе используем дефолт
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = defaultConfig
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exists: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return &cfg

}
