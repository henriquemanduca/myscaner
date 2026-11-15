package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"scaner-cli/internal/config"
	"scaner-cli/internal/logger"
	"scaner-cli/pkg/pdf"
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
	scan := &scanOptions{}
	a := &app{
		newConverter: func() (pdf.Converter, error) {
			return pdf.NewPopplerConverter()
		},
	}

	rootCmd := &cobra.Command{
		Use:   "scaner-cli",
		Short: "Extrai texto de imagens e PDFs via IA (OpenRouter ou LM Studio) em Markdown",
		Long: `scaner-cli - Extrai texto de imagens e documentos PDF via IA e salva o resultado
em formato Markdown (.md) no mesmo diretório e com o mesmo nome base do arquivo de entrada.
Suporta provedores de IA em nuvem (OpenRouter) ou modelos locais (LM Studio).
Para PDFs, converte cada página em imagem com alta resolução, envia para extração de OCR
e consolida todo o conteúdo em um único arquivo Markdown.`,
		Example: `  # Extrair texto via OpenRouter (padrão)
  scaner-cli -p /caminho/para/documento.png

  # Extrair texto usando LM Studio local (sem enviar dados para nuvem)
  scaner-cli -p /caminho/para/documento.png --provider lmstudio

  # LM Studio com modelo local e endpoint customizado
  scaner-cli -p ./relatorio.pdf --provider lmstudio --model qwen2.5-vl-7b-instruct --base-url http://localhost:1234/v1

  # PDF mantendo imagens temporárias das páginas
  scaner-cli -p ./contrato.pdf --keep-images`,
		Args:          cobra.NoArgs,
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Sem arquivo de entrada: exibe a ajuda em vez de falhar.
			if scan.inputPath == "" {
				return cmd.Help()
			}
			return a.runScan(cmd, scan)
		},
	}

	scan.bind(rootCmd)

	rootCmd.PersistentFlags().StringVarP(&flags.cfgFile, "config", "c", "", "caminho do arquivo de configuração")
	rootCmd.PersistentFlags().BoolVarP(&flags.verbose, "verbose", "v", false, "ativa logs de debug detalhados")
	rootCmd.PersistentFlags().StringVar(&flags.logFormat, "log-format", "", "formato de log (text ou json)")

	// Registra subcomandos
	rootCmd.AddCommand(newVersionCmd())

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
