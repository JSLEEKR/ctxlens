package pricing

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the user configuration file.
type Config struct {
	Providers map[string]map[string]ModelPricing `yaml:"providers"`
}

// DefaultConfigDir returns the default configuration directory.
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ctxlens")
}

// DefaultConfigPath returns the default configuration file path.
func DefaultConfigPath() string {
	dir := DefaultConfigDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "config.yaml")
}

// LoadConfig loads configuration from the given path.
// Returns nil config (no error) if the file doesn't exist.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// SaveDefaultConfig creates a default config file at the default path.
func SaveDefaultConfig() error {
	dir := DefaultConfigDir()
	if dir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	cfg := Config{
		Providers: map[string]map[string]ModelPricing{
			"anthropic": {
				"claude-sonnet-4": {Input: 3.0, Output: 15.0},
				"claude-opus-4":   {Input: 15.0, Output: 75.0},
			},
			"openai": {
				"gpt-4o": {Input: 2.5, Output: 10.0},
			},
		},
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
