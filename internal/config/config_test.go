package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("MY_SCANER_OPENROUTER_MODEL", "")
	t.Setenv("MY_SCANER_LOG_LEVEL", "")
	t.Setenv("MY_SCANER_LOG_FORMAT", "")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("expected no error loading default config, got %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected default LogLevel 'info', got '%s'", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("expected default LogFormat 'text', got '%s'", cfg.LogFormat)
	}
	if cfg.OpenRouter.Model != "google/gemini-2.0-flash-001" {
		t.Errorf("expected default model 'google/gemini-2.0-flash-001', got '%s'", cfg.OpenRouter.Model)
	}
}

func TestLoad_CustomFile(t *testing.T) {
	t.Setenv("MY_SCANER_OPENROUTER_MODEL", "")
	t.Setenv("MY_SCANER_LOG_LEVEL", "")
	t.Setenv("MY_SCANER_LOG_FORMAT", "")

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	content := []byte(`
log_level: "debug"
log_format: "json"
openrouter:
  api_key: "test-api-key"
  model: "anthropic/claude-3.5-sonnet"
`)
	if err := os.WriteFile(configFile, content, 0600); err != nil {
		t.Fatalf("failed to write temporary config: %v", err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("failed to load custom config: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got '%s'", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected LogFormat 'json', got '%s'", cfg.LogFormat)
	}
	if cfg.OpenRouter.APIKey != "test-api-key" {
		t.Errorf("expected APIKey 'test-api-key', got '%s'", cfg.OpenRouter.APIKey)
	}
	if cfg.OpenRouter.Model != "anthropic/claude-3.5-sonnet" {
		t.Errorf("expected model 'anthropic/claude-3.5-sonnet', got '%s'", cfg.OpenRouter.Model)
	}
}
