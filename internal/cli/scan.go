package cli

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"my-scaner/pkg/scanner"
)

func newScanCmd() *cobra.Command {
	var (
		host    string
		ports   []int
		timeout time.Duration
	)

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Executa varredura de portas em um host alvo",
		Example: `  # Varre portas padrão no localhost
  my-scaner scan --host localhost

  # Varre portas específicas com timeout customizado
  my-scaner scan --host 192.168.1.1 --ports 80,443,22 --timeout 1s`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Caso não tenham sido passadas portas via flag, usa do config se disponível
			if len(ports) == 0 && cfg != nil && len(cfg.Scanner.DefaultPorts) > 0 {
				ports = cfg.Scanner.DefaultPorts
			}
			if timeout == 0 && cfg != nil && cfg.Scanner.TimeoutSeconds > 0 {
				timeout = time.Duration(cfg.Scanner.TimeoutSeconds) * time.Second
			}

			slog.Debug("Iniciando varredura", "host", host, "ports", ports, "timeout", timeout)

			s := scanner.New()

			for _, port := range ports {
				select {
				case <-ctx.Done():
					slog.Warn("Varredura interrompida pelo usuário")
					return ctx.Err()
				default:
				}

				res := s.ScanPort(ctx, scanner.Target{Host: host, Port: port}, timeout)
				status := "CLOSED"
				if res.Open {
					status = "OPEN"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s:%d\n", status, host, port)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "localhost", "Host alvo da varredura")
	cmd.Flags().IntSliceVarP(&ports, "ports", "p", []int{80, 443, 8080}, "Portas a serem verificadas (separadas por vírgula)")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 2*time.Second, "Tempo limite por verificação de porta")

	return cmd
}
