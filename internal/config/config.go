package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port   string
	DBPath string
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		return Config{}, fmt.Errorf("DB_PATH environment variable is required")
	}

	return Config{
		Port:   port,
		DBPath: dbPath,
	}, nil
}
