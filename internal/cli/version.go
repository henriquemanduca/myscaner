package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"scaner-cli/internal/buildinfo"
)

func newVersionCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Exibe informações de versão e build da CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := buildinfo.Get()
			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}

			fmt.Fprintln(cmd.OutOrStdout(), info.String())
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "exibe a versão formatada em JSON")

	return cmd
}
