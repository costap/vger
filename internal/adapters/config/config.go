package config

import (
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// Config holds user-level vger settings persisted in ~/.vger/config.yaml.
type Config struct {
	// UserContext is a freeform description of the user's tech stack and environment.
	// When non-empty it is automatically prepended to all Gemini prompts so that
	// ask, research, and digest answers are tailored to the user's context.
	UserContext string `yaml:"user_context,omitempty"`
}

// DefaultPath returns ~/.vger/config.yaml.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".vger", "config.yaml"), nil
}

// Load reads the config file. If the file does not exist, a zero-value Config
// is returned without error — missing config is not an error condition.
func Load() (*Config, error) {
	path, err := DefaultPath()
	if err != nil {
		return &Config{}, nil
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes cfg to ~/.vger/config.yaml with 0600 permissions.
func Save(cfg *Config) error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}
