package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all configuration for the application
type Config struct {
	APIKey         string `json:"api_key"`
	APISecret      string `json:"api_secret"`
	APIHost        string `json:"api_host"`
	DefaultImageID string `json:"default_image_id"`
	LogLevel       string `json:"log_level"`
	Timeout        int    `json:"timeout_seconds"`
}

// Load reads the config file and environment variables to create a Config
func Load() (*Config, error) {
	config := &Config{
		LogLevel: "info",
		Timeout:  30,
	}

	// First, try to load from config file
	if err := loadFromFile(config); err != nil {
		fmt.Printf("Warning: Could not load config file: %v\n", err)
	}

	// Then, override with environment variables
	loadFromEnv(config)

	// Validate the config
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func loadFromFile(config *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".kasm-stress-test.json")
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("could not decode config file: %w", err)
	}

	return nil
}

func loadFromEnv(config *Config) {
	if key := os.Getenv("KASM_KEY"); key != "" {
		config.APIKey = key
	}
	if secret := os.Getenv("KASM_SECRET"); secret != "" {
		config.APISecret = secret
	}
	if host := os.Getenv("KASM_API_HOST"); host != "" {
		config.APIHost = host
	}
	if imageID := os.Getenv("KASM_DEFAULT_IMAGE_ID"); imageID != "" {
		config.DefaultImageID = imageID
	}
	if logLevel := os.Getenv("KASM_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}
	if timeout := os.Getenv("KASM_TIMEOUT"); timeout != "" {
		// Parse timeout to int and set if valid
	}
}

func (c *Config) validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.APISecret == "" {
		return fmt.Errorf("API secret is required")
	}
	if c.APIHost == "" {
		return fmt.Errorf("API host is required")
	}
	return nil
}
