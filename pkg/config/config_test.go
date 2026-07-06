package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadDefaultsWhenNoFile(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	def := Default()
	if !reflect.DeepEqual(cfg.Engines.Enabled, def.Engines.Enabled) {
		t.Errorf("engines = %v, want default %v", cfg.Engines.Enabled, def.Engines.Enabled)
	}
	if cfg.LLM.Model != def.LLM.Model || cfg.Limit != def.Limit {
		t.Errorf("defaults not applied: %+v", cfg)
	}
}

func TestLoadYAMLOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	yaml := `
llm:
  base_url: "https://openrouter.ai/api/v1"
  model: "anthropic/claude-3-haiku"
  temperature: 0.2
engines:
  enabled: [fofa, censys]
output:
  format: json
  color: never
limit: 25
`
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLM.BaseURL != "https://openrouter.ai/api/v1" || cfg.LLM.Model != "anthropic/claude-3-haiku" {
		t.Errorf("llm not overridden: %+v", cfg.LLM)
	}
	if !reflect.DeepEqual(cfg.Engines.Enabled, []string{"fofa", "censys"}) {
		t.Errorf("engines not overridden: %v", cfg.Engines.Enabled)
	}
	if cfg.Output.Format != "json" || cfg.Output.Color != "never" || cfg.Limit != 25 {
		t.Errorf("output/limit not overridden: %+v", cfg)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("llm:\n  model: from-file\n  api_key: file-key\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvLLMModel, "from-env")
	t.Setenv(EnvLLMAPIKey, "env-key")
	t.Setenv(EnvLLMBaseURL, "https://env.example/v1")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLM.Model != "from-env" {
		t.Errorf("env should override file model, got %q", cfg.LLM.Model)
	}
	if cfg.LLM.APIKey != "env-key" {
		t.Errorf("env should override file api_key, got %q", cfg.LLM.APIKey)
	}
	if cfg.LLM.BaseURL != "https://env.example/v1" {
		t.Errorf("env should override base_url, got %q", cfg.LLM.BaseURL)
	}
}

func TestInvalidYAMLErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("llm: [this is not a map\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("expected error for invalid YAML")
	}
}
