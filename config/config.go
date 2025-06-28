package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/sanity-io/litter"
)

type (
	Config struct {
		Name string `env:"APP_NAME,required"`
		Port string `env:"HTTP_PORT,required" envDefault:":8080"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it. Falling back to system environment variables.")
		// Dont fatal here, allow the app to run with just system env vars.
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	cfg.Port = fmt.Sprintf(":%s", cfg.Port)

	fmt.Println("Configured successfully, ", litter.Sdump(cfg))
	return cfg, nil
}
