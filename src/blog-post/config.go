package blogpost

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

const (
	hatenaId     = "GABA_FORTNITE"
	hatenaBlogId = "gaba-fortnite.hatenablog.com"
)

type Config struct {
	OpenAIAPIKey string `json:"openai_api_key"`
	HatenaId     string `json:"hatena_id"`
	HatenaBlogId string `json:"hatena_blog_id"`
	HatenaApiKey string `json:"hatena_api_key"`
}

func loadFromEnv() *Config {
	// Environment variable values from GCP Secret Manager
	// If the environment variable value starts with "sm://", it is treated as an automatically retrieved value from Secret Manager
	// HatenaId and HatenaBlogId are defined as fixed values
	config := &Config{
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		HatenaId:     hatenaId,
		HatenaBlogId: hatenaBlogId,
		HatenaApiKey: os.Getenv("HATENA_API_KEY"),
	}

	// Verify that required configuration values are specified
	if config.OpenAIAPIKey != "" && config.HatenaApiKey != "" {
		return config
	}
	return nil
}

func loadConfig() (*Config, error) {
	// 1. Load configuration from environment variables
	config := loadFromEnv()
	if config != nil {
		return config, nil
	}

	// 2. If environment variable configuration is incomplete, load from configuration file
	configFile := "config.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		// If configuration file reading fails, output a warning log but do not abort
		slog.Warn("Failed to read config file", "error", err)

		// If configuration cannot be obtained from both environment variables and configuration file, return an error
		if config == nil {
			return nil, fmt.Errorf("failed to load configuration from environment variables or config file: %v", err)
		}
		return config, nil
	}

	var fileConfig Config
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		slog.Error("Failed to parse JSON", "error", err)

		// If configuration cannot be obtained from both environment variables and configuration file, return an error
		if config == nil {
			return nil, fmt.Errorf("failed to parse JSON: %v", err)
		}
		return config, nil
	}

	return &fileConfig, nil
}
