package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env:"ENV" env-default:"local"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Database   Database   `yaml:"database"`
}

type HTTPServer struct {
	Address         string        `yaml:"address"      env:"HTTP_ADDRESS" env-default:"localhost"`
	HTTPPort        string        `yaml:"port"         env:"HTTP_PORT"    env-default:"8080"`
	Timeout         time.Duration `yaml:"timeout"      env-default:"4s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"  env:"HTTP_SHUTDOWN_TIMEOUT"  env-default:"10s"`
}

type Database struct {
	Driver   string  `yaml:"driver"   env:"DB_DRIVER"   env-default:"postgres"` // postgres | mysql - задел на будущее
	Host     string  `yaml:"host"     env:"DB_HOST"     env-default:"localhost"`
	Port     string  `yaml:"port"     env:"DB_PORT"     env-default:"5432"`
	User     string  `yaml:"user"     env:"DB_USER"     env-default:"postgres"`
	Password string  `yaml:"password" env:"DB_PASSWORD" env-default:"postgres"`
	Name     string  `yaml:"name"     env:"DB_NAME"     env-default:"subscriptions"`
	SSLMode  string  `yaml:"sslmode"  env:"DB_SSLMODE"  env-default:"disable"`
	Pool     *DBPool `yaml:"pool,omitempty"` // nil, если секции database.pool нет
}

type DBPool struct {
	MaxOpenConns int           `yaml:"max_open_conns"  env:"DB_POOL_MAX_OPEN_CONNS"  env-default:"10"`
	MaxIdleConns int           `yaml:"max_idle_conns"  env:"DB_POOL_MAX_IDLE_CONNS"  env-default:"5"`
	ConnLifetime time.Duration `yaml:"conn_lifetime"   env:"DB_POOL_CONN_LIFETIME"   env-default:"5m"`
}

const defaultConfig = "./config/config.yaml"

func LoadConfig() *Config {
	// 1. Путь из ENV, иначе дефолт
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = defaultConfig
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	// ReadConfig читает YAML/JSON/TOML по расширению и применяет ENV-переменные согласно тегам.
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	return &cfg
}
