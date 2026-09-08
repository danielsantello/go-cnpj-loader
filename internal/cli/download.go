package cli

import (
	"fmt"
	"net/http"
	"time"

	"github.com/danielsantello/go-cnpj-loader/internal/download"
	"github.com/spf13/cobra"
)

func newDownloadCommand() *cobra.Command {
	var (
		destinationDirectory string
		referenceYear        uint16
		referenceMonth       uint8
	)

	command := &cobra.Command{
		Use:   "download",
		Short: "Baixa uma publicação da Receita Federal",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			transport := http.DefaultTransport.(*http.Transport).Clone()
			transport.ResponseHeaderTimeout = 60 * time.Second

			client := &http.Client{
				Transport: transport,
			}

			fmt.Fprintln(
				command.OutOrStdout(),
				"Consultando os arquivos da publicação...",
			)

			result, err := download.DownloadPublication(
				command.Context(),
				client,
				referenceYear,
				referenceMonth,
				destinationDirectory,
				func(
					current int,
					total int,
					file download.File,
				) {
					fmt.Fprintf(
						command.OutOrStdout(),
						"Processando arquivo %d/%d: %s (%d bytes)...\n",
						current,
						total,
						file.Name,
						file.SizeBytes,
					)
				},
			)
			if err != nil {
				return fmt.Errorf(
					"não foi possível baixar a publicação: %w",
					err,
				)
			}

			fmt.Fprintf(
				command.OutOrStdout(),
				"Publicação disponível em %q: %d arquivos, "+
					"%d baixados, %d reaproveitados, %d bytes.\n",
				destinationDirectory,
				result.FileCount,
				result.DownloadedCount,
				result.ReusedCount,
				result.TotalSizeBytes,
			)

			return nil
		},
	}

	command.Flags().StringVar(
		&destinationDirectory,
		"destination",
		"",
		"diretório absoluto que receberá os arquivos ZIP",
	)

	command.Flags().Uint16Var(
		&referenceYear,
		"reference-year",
		0,
		"ano de referência da publicação",
	)

	command.Flags().Uint8Var(
		&referenceMonth,
		"reference-month",
		0,
		"mês de referência da publicação",
	)

	return command
}
