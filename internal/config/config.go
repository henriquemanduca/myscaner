package config

import (
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
	OpenRouter OpenRouterConfig `mapstructure:"openrouter"`
}

// OpenRouterConfig contém os parâmetros de conexão com a API do OpenRouter.
type OpenRouterConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// Load carrega configurações a partir de arquivo e/ou variáveis de ambiente.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	// Valores padrão (sensible defaults)
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "text")
	v.SetDefault("openrouter.base_url", "https://openrouter.ai/api/v1")
	v.SetDefault("openrouter.model", "google/gemini-2.0-flash-001")

	// Suporte a variáveis de ambiente com prefixo: MY_SCANER_OPENROUTER_API_KEY, etc.
	v.SetEnvPrefix("MY_SCANER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(filepath.Join(home, ".config", "my-scaner"))
		}
		v.AddConfigPath("$HOME/.config/my-scaner")
		v.AddConfigPath("/etc/my-scaner")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && cfgFile != "" {
			return nil, fmt.Errorf("falha ao ler arquivo de configuração: %w", err)
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
