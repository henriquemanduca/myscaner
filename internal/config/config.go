package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config representa as opções de configuração do aplicativo.
type Config struct {
	LogLevel  string        `mapstructure:"log_level"`
	LogFormat string        `mapstructure:"log_format"`
	Scanner   ScannerConfig `mapstructure:"scanner"`
}

// ScannerConfig contém os parâmetros específicos do módulo scanner.
type ScannerConfig struct {
	TimeoutSeconds int   `mapstructure:"timeout_seconds"`
	DefaultPorts   []int `mapstructure:"default_ports"`
	Concurrency    int   `mapstructure:"concurrency"`
}

// Load carrega configurações a partir de arquivo e/ou variáveis de ambiente.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	// Valores padrão (sensible defaults)
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "text")
	v.SetDefault("scanner.timeout_seconds", 3)
	v.SetDefault("scanner.default_ports", []int{80, 443, 8080})
	v.SetDefault("scanner.concurrency", 10)

	// Suporte a variáveis de ambiente: MY_SCANER_LOG_LEVEL, etc.
	v.SetEnvPrefix("MY_SCANER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/my-scaner")
		v.AddConfigPath("/etc/my-scaner")
	}

	if err := v.ReadInConfig(); err != nil {
		// Se o usuário especificou explicitamente um arquivo e ele não existe, retorna erro
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && cfgFile != "" {
			return nil, fmt.Errorf("falha ao ler arquivo de configuração: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("falha ao decodificar configurações: %w", err)
	}

	return &cfg, nil
}
