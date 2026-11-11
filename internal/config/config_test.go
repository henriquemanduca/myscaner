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
	t.Setenv("MY_SCANER_PROVIDER", "")

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
	if cfg.Provider != "openrouter" {
		t.Errorf("expected default provider 'openrouter', got '%s'", cfg.Provider)
	}
	if cfg.OpenRouter.Model != "google/gemini-2.0-flash-001" {
		t.Errorf("expected default model 'google/gemini-2.0-flash-001', got '%s'", cfg.OpenRouter.Model)
	}
	if cfg.LMStudio.BaseURL != "http://localhost:1234/v1" {
		t.Errorf("expected default LMStudio BaseURL 'http://localhost:1234/v1', got '%s'", cfg.LMStudio.BaseURL)
	}
}

func TestLoad_RejectsJSONFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	content := []byte(`{"provider": "lmstudio"}`)
	if err := os.WriteFile(configFile, content, 0600); err != nil {
		t.Fatalf("failed to write temporary json config: %v", err)
	}

	if _, err := Load(configFile); err == nil {
		t.Fatal("expected error loading json config, got nil")
	}
}

func TestLoad_CustomYAMLFile(t *testing.T) {
	t.Setenv("MY_SCANER_OPENROUTER_MODEL", "")
	t.Setenv("MY_SCANER_LOG_LEVEL", "")
	t.Setenv("MY_SCANER_LOG_FORMAT", "")
	t.Setenv("MY_SCANER_PROVIDER", "")

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	content := []byte(`
log_level: "debug"
log_format: "json"
provider: "lmstudio"
openrouter:
  api_key: "test-api-key"
  model: "anthropic/claude-3.5-sonnet"
lmstudio:
  base_url: "http://192.168.1.50:1234/v1"
  model: "qwen2.5-vl-7b-instruct"
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
	if cfg.Provider != "lmstudio" {
		t.Errorf("expected Provider 'lmstudio', got '%s'", cfg.Provider)
	}
	if cfg.OpenRouter.APIKey != "test-api-key" {
		t.Errorf("expected APIKey 'test-api-key', got '%s'", cfg.OpenRouter.APIKey)
	}
	if cfg.OpenRouter.Model != "anthropic/claude-3.5-sonnet" {
		t.Errorf("expected model 'anthropic/claude-3.5-sonnet', got '%s'", cfg.OpenRouter.Model)
	}
	if cfg.LMStudio.BaseURL != "http://192.168.1.50:1234/v1" {
		t.Errorf("expected LMStudio BaseURL 'http://192.168.1.50:1234/v1', got '%s'", cfg.LMStudio.BaseURL)
	}
	if cfg.LMStudio.Model != "qwen2.5-vl-7b-instruct" {
		t.Errorf("expected LMStudio Model 'qwen2.5-vl-7b-instruct', got '%s'", cfg.LMStudio.Model)
	}
}

func TestDefaultConfigFilePath(t *testing.T) {
	path, err := DefaultConfigFilePath()
	if err != nil {
		t.Fatalf("expected no error getting default config path, got %v", err)
	}
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected file name 'config.yaml', got '%s'", filepath.Base(path))
	}
}
