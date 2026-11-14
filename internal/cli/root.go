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
	"my-scaner/pkg/pdf"
)

type globalFlags struct {
	cfgFile   string
	verbose   bool
	logFormat string
}

type app struct {
	cfg          *config.Config
	log          *slog.Logger
	newConverter func() (pdf.Converter, error)
}

func newRootCmd() *cobra.Command {
	var flags globalFlags
	a := &app{
		newConverter: func() (pdf.Converter, error) {
			return pdf.NewPopplerConverter()
		},
	}

	rootCmd := &cobra.Command{
		Use:   "my-scaner",
		Short: "my-scaner é uma ferramenta CLI moderna e extensível em Go",
		Long: `my-scaner - Utilitário de linha de comando modular e de alta performance,
estruturado segundo os padrões de design e boas práticas do ecossistema Go.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.cfgFile)
			if err != nil {
				return fmt.Errorf("erro ao inicializar configuração: %w", err)
			}
			a.cfg = cfg

			logLevel := cfg.LogLevel
			if flags.verbose {
				logLevel = "debug"
			}

			format := flags.logFormat
			if format == "" {
				format = cfg.LogFormat
			}

			a.log = logger.New(logger.Config{
				Level:  logLevel,
				Format: format,
				Output: os.Stderr,
			})
			slog.SetDefault(a.log)

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&flags.cfgFile, "config", "c", "", "caminho do arquivo de configuração")
	rootCmd.PersistentFlags().BoolVarP(&flags.verbose, "verbose", "v", false, "ativa logs de debug detalhados")
	rootCmd.PersistentFlags().StringVar(&flags.logFormat, "log-format", "", "formato de log (text ou json)")

	// Registra subcomandos
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newScanCmd(a))

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
