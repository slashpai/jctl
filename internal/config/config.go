package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	BaseURL string `yaml:"base_url"`
	Email   string `yaml:"email"`
	Token   string `yaml:"token"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "jctl"), nil
}

func ConfigFilePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// LoadOption configures the behavior of Load.
type LoadOption func(*loadOptions)

type loadOptions struct {
	useEnv bool
}

// WithEnv controls whether environment variables override config file values.
// Defaults to true.
func WithEnv(enabled bool) LoadOption {
	return func(o *loadOptions) {
		o.useEnv = enabled
	}
}

func Load(opts ...LoadOption) (*Config, error) {
	o := &loadOptions{useEnv: true}
	for _, opt := range opts {
		opt(o)
	}

	path, err := ConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("finding config directory: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found — run `jctl configure` first")
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if o.useEnv {
		if v := os.Getenv("JCTL_BASE_URL"); v != "" {
			cfg.BaseURL = v
		}
		if v := os.Getenv("JCTL_EMAIL"); v != "" {
			cfg.Email = v
		}
		if v := os.Getenv("JCTL_TOKEN"); v != "" {
			cfg.Token = v
		}
	}

	if cfg.BaseURL == "" || cfg.Email == "" || cfg.Token == "" {
		return nil, fmt.Errorf("incomplete config — run `jctl configure` to set base_url, email, and token")
	}

	if !strings.HasPrefix(cfg.BaseURL, "https://") {
		return nil, fmt.Errorf("base_url must use HTTPS to protect credentials in transit")
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return fmt.Errorf("finding config directory: %w", err)
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
