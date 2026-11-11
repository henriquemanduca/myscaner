package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config representa as opções de configuração do aplicativo.
type Config struct {
	LogLevel   string           `mapstructure:"log_level"`
	LogFormat  string           `mapstructure:"log_format"`
	Provider   string           `mapstructure:"provider"` // "openrouter" ou "lmstudio"
	OpenRouter OpenRouterConfig `mapstructure:"openrouter"`
	LMStudio   LMStudioConfig   `mapstructure:"lmstudio"`
}

// OpenRouterConfig contém os parâmetros de conexão com a API do OpenRouter.
type OpenRouterConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// LMStudioConfig contém os parâmetros de conexão com o LM Studio local.
type LMStudioConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
	APIKey  string `mapstructure:"api_key"` // Opcional no LM Studio local
}

// DefaultConfigDir retorna o caminho padrão para configurações (~/.config/my-scaner).
func DefaultConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "my-scaner"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("não foi possível determinar o diretório home do usuário: %w", err)
	}
	return filepath.Join(home, ".config", "my-scaner"), nil
}

// DefaultConfigFilePath retorna o caminho completo padrão do arquivo YAML (~/.config/my-scaner/config.yaml).
func DefaultConfigFilePath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load carrega configurações a partir de arquivo YAML e/ou variáveis de ambiente.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	// Valores padrão (sensible defaults)
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "text")
	v.SetDefault("provider", "openrouter")

	// OpenRouter defaults
	v.SetDefault("openrouter.base_url", "https://openrouter.ai/api/v1")
	v.SetDefault("openrouter.model", "google/gemini-2.0-flash-001")

	// LM Studio defaults
	v.SetDefault("lmstudio.base_url", "http://localhost:1234/v1")
	v.SetDefault("lmstudio.model", "default")
	v.SetDefault("lmstudio.api_key", "")

	// Suporte a variáveis de ambiente com prefixo: MY_SCANER_OPENROUTER_API_KEY, MY_SCANER_PROVIDER, etc.
	v.SetEnvPrefix("MY_SCANER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if cfgFile != "" {
		ext := strings.ToLower(filepath.Ext(cfgFile))
		if ext != ".yaml" && ext != ".yml" {
			return nil, fmt.Errorf("formato de configuração não suportado '%s': apenas arquivos YAML (.yaml/.yml) são aceitos", ext)
		}
		v.SetConfigFile(cfgFile)
		v.SetConfigType("yaml")
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("falha ao ler arquivo de configuração '%s': %w", cfgFile, err)
		}
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		addConfigPaths(v)

		if err := v.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if !errors.As(err, &notFound) {
				return nil, fmt.Errorf("falha ao ler arquivo de configuração: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("falha ao decodificar configurações: %w", err)
	}

	// Fallback para a variável padrão do sistema OPENROUTER_API_KEY caso não configurada via arquivo ou MY_SCANER_
	if cfg.OpenRouter.APIKey == "" {
		cfg.OpenRouter.APIKey = os.Getenv("OPENROUTER_API_KEY")
	}

	return &cfg, nil
}

func addConfigPaths(v *viper.Viper) {
	v.AddConfigPath(".")
	if dir, err := DefaultConfigDir(); err == nil {
		v.AddConfigPath(dir)
	}
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "my-scaner"))
	}
	v.AddConfigPath("/etc/my-scaner")
}
