package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"my-scaner/internal/config"
	"my-scaner/internal/logger"
)

type globalFlags struct {
	cfgFile   string
	verbose   bool
	logFormat string
}

var (
	flags globalFlags
	cfg   *config.Config
	log   *slog.Logger
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "my-scaner",
		Short: "my-scaner é uma ferramenta CLI moderna e extensível em Go",
		Long: `my-scaner - Utilitário de linha de comando modular e de alta performance,
estruturado segundo os padrões de design e boas práticas do ecossistema Go.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cfg, err = config.Load(flags.cfgFile)
			if err != nil {
				return fmt.Errorf("erro ao inicializar configuração: %w", err)
			}

			logLevel := cfg.LogLevel
			if flags.verbose {
				logLevel = "debug"
			}

			format := flags.logFormat
			if format == "" {
				format = cfg.LogFormat
			}

			log = logger.New(logger.Config{
				Level:  logLevel,
				Format: format,
				Output: os.Stderr,
			})
			slog.SetDefault(log)

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&flags.cfgFile, "config", "c", "", "caminho do arquivo de configuração")
	rootCmd.PersistentFlags().BoolVarP(&flags.verbose, "verbose", "v", false, "ativa logs de debug detalhados")
	rootCmd.PersistentFlags().StringVar(&flags.logFormat, "log-format", "", "formato de log (text ou json)")

	// Registra subcomandos
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newScanCmd())

	return rootCmd
}

// Execute executa o comando raiz com suporte a graceful shutdown (SIGINT/SIGTERM).
func Execute() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cmd := newRootCmd()
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		return err
	}
	return nil
}
