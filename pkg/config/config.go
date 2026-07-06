// Package config loads uncover-turbo configuration from a YAML file with
// environment-variable overrides. Precedence (lowest to highest) is:
//
//	built-in defaults < YAML file < environment variables < CLI flags
//
// This package covers defaults, file and env; CLI flags are layered on top by
// the caller (cmd/uncover-turbo).
//
// Engine API credentials are intentionally NOT managed here — they continue to
// use uncover's own provider mechanism (env vars / provider-config). This config
// only governs the LLM backend, engine enable/disable and output preferences.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LLM configures the backend LLM used for the natural-language -> IR upper layer.
// BaseURL makes it work with OpenAI, OpenRouter or any OpenAI-compatible endpoint.
type LLM struct {
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	APIKey      string  `yaml:"api_key"`
	Temperature float32 `yaml:"temperature"`
}

// Engines configures which recon engines are queried.
type Engines struct {
	Enabled []string `yaml:"enabled"`
}

// Output configures rendering.
type Output struct {
	Format string `yaml:"format"` // text | json
	Color  string `yaml:"color"`  // auto | always | never
}

// Config is the full configuration.
type Config struct {
	LLM     LLM     `yaml:"llm"`
	Engines Engines `yaml:"engines"`
	Output  Output  `yaml:"output"`
	Limit   int     `yaml:"limit"`
	Proxy   string  `yaml:"proxy"`
}

// Environment variable names for sensitive / commonly-overridden LLM settings.
const (
	EnvLLMAPIKey  = "UNCOVER_TURBO_LLM_API_KEY"
	EnvLLMBaseURL = "UNCOVER_TURBO_LLM_BASE_URL"
	EnvLLMModel   = "UNCOVER_TURBO_LLM_MODEL"
)

// Default returns the built-in default configuration.
func Default() *Config {
	return &Config{
		LLM: LLM{
			BaseURL:     "https://api.openai.com/v1",
			Model:       "gpt-3.5-turbo",
			Temperature: 0.0,
		},
		Engines: Engines{Enabled: []string{"fofa", "censys", "hunter", "zoomeye"}},
		Output:  Output{Format: "text", Color: "auto"},
		Limit:   100,
	}
}

// DefaultPath returns the default config file location
// (~/.config/uncover-turbo/config.yaml). It returns "" if the home directory
// cannot be determined.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "uncover-turbo", "config.yaml")
}

// Load builds a Config by layering, in increasing precedence: defaults, the YAML
// file at path (if it exists), then environment variables. A missing file is not
// an error — defaults are used. Pass an empty path to skip file loading entirely;
// callers decide which path to use (typically DefaultPath).
func Load(path string) (*Config, error) {
	cfg := Default()

	if path != "" {
		data, err := os.ReadFile(path)
		switch {
		case err == nil:
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("config: parsing %s: %w", path, err)
			}
		case os.IsNotExist(err):
			// Missing file: fall back to defaults, not an error.
		default:
			return nil, fmt.Errorf("config: reading %s: %w", path, err)
		}
	}

	applyEnv(cfg)
	return cfg, nil
}

// applyEnv overrides LLM fields from environment variables when set to a
// non-empty value. An empty variable is ignored so it never clears a value that
// came from the config file.
func applyEnv(cfg *Config) {
	if v, ok := os.LookupEnv(EnvLLMAPIKey); ok && v != "" {
		cfg.LLM.APIKey = v
	}
	if v, ok := os.LookupEnv(EnvLLMBaseURL); ok && v != "" {
		cfg.LLM.BaseURL = v
	}
	if v, ok := os.LookupEnv(EnvLLMModel); ok && v != "" {
		cfg.LLM.Model = v
	}
}
