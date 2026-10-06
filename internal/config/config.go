package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string
	DatabaseURL string
	CatalogURL  string
}

func Load() (Config, error) {
	if os.Getenv("ENV") != "prod" {
		_ = godotenv.Load()
	}
	cfg := Config{
		Port:        os.Getenv("PORT"),
		Env:         os.Getenv("ENV"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		CatalogURL:  os.Getenv("CATALOG_SERVICE_URL"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.CatalogURL == "" {
		return Config{}, fmt.Errorf("CATALOG_SERVICE_URL is required")
	}
	return cfg, nil
}
