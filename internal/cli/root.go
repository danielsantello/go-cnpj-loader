package cli

import (
	"github.com/danielsantello/go-cnpj-loader/internal/buildinfo"
	"github.com/spf13/cobra"
	"io"
)

func NewRootCommand(output io.Writer, errorOutput io.Writer) *cobra.Command {
	command := &cobra.Command{
		Use:           "cnpj-loader",
		Short:         "Carrega os dados públicos de CNPJ no MySQL",
		Long:          "cnpj-loader é um carregador versionado dos dados públicos de CNPJ disponibilizados pela Receita Federal.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}

	command.SetOut(output)
	command.SetErr(errorOutput)

	command.AddCommand(newVersionCommand(buildinfo.Current()))
	command.AddCommand(newMigrateControlCommand(buildinfo.Current()))

	return command
}
