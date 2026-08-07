package utils

import (
"encoding/json"
"os"
"sync"
)

// Config holds the application configuration loaded from config.json or env vars.
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
}

// ServerConfig holds server-related config.
type ServerConfig struct {
	Port    string `json:"port"`
	BaseURL string `json:"baseUrl"`
}

// DatabaseConfig holds database-related config.
type DatabaseConfig struct {
	URL             string `json:"url"`
	MaxOpenConns    int    `json:"maxOpenConns"`
	MaxIdleConns    int    `json:"maxIdleConns"`
	ConnMaxLifetime int    `json:"connMaxLifetimeSeconds"`
}

var (
config     *Config
configOnce sync.Once
)

// GetConfig returns the singleton config instance.
func GetConfig() *Config {
	configOnce.Do(func() {
		config = loadConfig()
	})
	return config
}

func loadConfig() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Port:    "8080",
			BaseURL: "http://localhost:8080",
		},
		Database: DatabaseConfig{
			URL:             "postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 300,
		},
	}

	// Try loading from config.json
	if data, err := os.ReadFile("config.json"); err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	// Env vars override config.json
	if v := os.Getenv("PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		cfg.Server.BaseURL = v
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.Database.URL = v
	}

	return cfg
}
