package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/chihqiang/dbxgo/output"
	"github.com/chihqiang/dbxgo/source"
	"github.com/chihqiang/dbxgo/store"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func init() {
	_ = godotenv.Load()
}

// Config defines the global configuration structure
// It is used to load all configuration items for the application from the configuration file
type Config struct {
	Store  store.Config  `yaml:"store" json:"store" mapstructure:"store"`
	Source source.Config `yaml:"source" json:"source" mapstructure:"source"`
	Output output.Config `yaml:"output" json:"output" mapstructure:"output"`
}

// Load attempts to load the configuration.
// Load order: prioritizes reading from the file → if the file does not exist or fails to parse, it loads from environment variables.
func Load(path string) (*Config, error) {
	var cfg Config

	// ① Try loading from the configuration file first
	data, err := os.ReadFile(path)
	if err == nil {
		// Attempt to parse the YAML configuration file
		if yamlErr := yaml.Unmarshal(data, &cfg); yamlErr == nil {
			return &cfg, nil // File read and parsed successfully
		}
		// If YAML parsing fails, continue trying environment variables
	}

	// ② If file loading fails, try loading from environment variables
	if envErr := env.Parse(&cfg); envErr == nil {
		return &cfg, nil // Environment variable parsing successful
	}

	// ③ If both methods fail, return an error message
	return nil, fmt.Errorf("failed to load configuration (file: %v, env: %v)", err, env.Parse(&cfg))
}
