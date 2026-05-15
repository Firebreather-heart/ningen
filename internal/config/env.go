package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port        string
	LLMProvider string // e.g., "openai", "anthropic", "google"
	LLMAPIKey   string
	LLMModel    string
	DatabaseURI string
}

// Load reads configuration from environment variables and returns a Config.
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		LLMProvider: getEnv("LLM_PROVIDER", "openai"),
		LLMAPIKey:   os.Getenv("LLM_API_KEY"),
		LLMModel:    getEnv("LLM_MODEL", "gpt-4"),
		DatabaseURI: os.Getenv("DATABASE_URI"),
	}

	if cfg.LLMAPIKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY environment variable is required")
	}

	return cfg, nil
}

// getEnv returns the value of an environment variable or a default value if not set.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
