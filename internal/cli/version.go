package cli

import (
	"fmt"

	"github.com/danielsantello/go-cnpj-loader/internal/buildinfo"
	"github.com/spf13/cobra"
)

func newVersionCommand(info buildinfo.Info) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Exibe informações da versão instalada",
		Args:  cobra.NoArgs,
		Run: func(command *cobra.Command, _ []string) {
			output := command.OutOrStdout()

			fmt.Fprintf(output, "Versão: %s\n", info.Version)
			fmt.Fprintf(output, "Commit: %s\n", info.Commit)
			fmt.Fprintf(output, "Compilado em: %s\n", info.BuildDate)
			fmt.Fprintf(output, "Go: %s\n", info.GoVersion)
		},
	}
}
