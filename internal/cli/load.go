package cli

import (
	"context"
	"fmt"

	"github.com/danielsantello/go-cnpj-loader/internal/buildinfo"
	"github.com/danielsantello/go-cnpj-loader/internal/config"
	"github.com/danielsantello/go-cnpj-loader/internal/control"
	"github.com/danielsantello/go-cnpj-loader/internal/control/migrations"
	"github.com/danielsantello/go-cnpj-loader/internal/database"
	"github.com/danielsantello/go-cnpj-loader/internal/publication"
	"github.com/spf13/cobra"
)

type loadResult struct {
	publicationID uint64
	versionID     uint64
	fileCount     int
	countryCount  uint64
	controlSchema string
	dataSchema    string
}

func newLoadCommand(info buildinfo.Info) *cobra.Command {
	var (
		sourceValue    string
		referenceYear  uint16
		referenceMonth uint8
	)

	command := &cobra.Command{
		Use:   "load",
		Short: "Carrega uma publicação dos dados públicos de CNPJ",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			fmt.Fprintln(
				command.OutOrStdout(),
				"Descobrindo e verificando os arquivos da publicação...",
			)

			result, err := loadDirectoryPublication(
				command.Context(),
				info,
				sourceValue,
				referenceYear,
				referenceMonth,
				func(
					current int,
					total int,
					file publication.ClassifiedFile,
				) {
					fmt.Fprintf(
						command.OutOrStdout(),
						"Verificando arquivo %d/%d: %s\n",
						current,
						total,
						file.SourceName,
					)
				},
			)

			if err != nil {
				return err
			}

			fmt.Fprintf(
				command.OutOrStdout(),
				"Publicação %d registrada com %d arquivos no schema %q.\n",
				result.publicationID,
				result.fileCount,
				result.controlSchema,
			)

			fmt.Fprintf(
				command.OutOrStdout(),
				"Versão %d criada no schema %q com %d países carregados.\n",
				result.versionID,
				result.dataSchema,
				result.countryCount,
			)

			return nil
		},
	}

	command.Flags().StringVar(
		&sourceValue,
		"source",
		"",
		"diretório que contém os arquivos ZIP da publicação",
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

func loadDirectoryPublication(
	ctx context.Context,
	info buildinfo.Info,
	sourceValue string,
	referenceYear uint16,
	referenceMonth uint8,
	progress publication.VerifyFileProgress,
) (loadResult, error) {
	if referenceYear < 1000 || referenceYear > 9999 {
		return loadResult{}, fmt.Errorf(
			"ano de referência é inválido: %d",
			referenceYear,
		)
	}

	if referenceMonth < 1 || referenceMonth > 12 {
		return loadResult{}, fmt.Errorf(
			"mês de referência é inválido: %d",
			referenceMonth,
		)
	}

	source, err := publication.ParseSource(sourceValue)
	if err != nil {
		return loadResult{}, fmt.Errorf(
			"origem da publicação é inválida: %w",
			err,
		)
	}

	if source.Type != publication.SourceTypeDirectory {
		return loadResult{}, fmt.Errorf(
			"o comando load atualmente aceita somente diretórios",
		)
	}

	value, err := config.Load()
	if err != nil {
		return loadResult{}, fmt.Errorf(
			"não foi possível carregar a configuração: %w",
			err,
		)
	}

	if err := config.ValidateLoad(value); err != nil {
		return loadResult{}, fmt.Errorf(
			"configuração inválida: %w",
			err,
		)
	}

	catalog, err := publication.LoadCatalog()
	if err != nil {
		return loadResult{}, fmt.Errorf(
			"não foi possível carregar o catálogo de publicações: %w",
			err,
		)
	}

	files, err := publication.DiscoverDirectoryPublicationWithProgress(
		catalog,
		source,
		progress,
	)
	if err != nil {
		return loadResult{}, err
	}

	connection, err := database.OpenMySQL(value.MySQL)
	if err != nil {
		return loadResult{}, fmt.Errorf(
			"não foi possível abrir o gerenciador de conexões: %w",
			err,
		)
	}
	defer connection.Close()

	pingContext, cancel := context.WithTimeout(
		ctx,
		value.MySQL.ConnectTimeout,
	)
	defer cancel()

	if err := database.PingMySQL(
		pingContext,
		connection,
	); err != nil {
		return loadResult{}, err
	}

	if err := migrations.Migrate(
		ctx,
		connection,
		value.ControlSchema,
		info,
	); err != nil {
		return loadResult{}, fmt.Errorf(
			"não foi possível migrar o schema de controle: %w",
			err,
		)
	}

	publicationID, err := control.RegisterAvailablePublication(
		ctx,
		connection,
		value.ControlSchema,
		control.RegisterPublicationInput{
			ReferenceYear:  referenceYear,
			ReferenceMonth: referenceMonth,
			Source:         source,
			Files:          files,
		},
	)
	if err != nil {
		return loadResult{}, fmt.Errorf(
			"não foi possível registrar a publicação: %w",
			err,
		)
	}

	version, countryCount, err := createVersionAndLoadCountries(
		ctx,
		connection,
		value,
		publicationID,
		referenceYear,
		referenceMonth,
		files,
	)
	if err != nil {
		return loadResult{}, fmt.Errorf(
			"não foi possível iniciar a carga da versão: %w",
			err,
		)
	}

	return loadResult{
		publicationID: publicationID,
		versionID:     version.ID,
		fileCount:     len(files),
		countryCount:  countryCount,
		controlSchema: value.ControlSchema,
		dataSchema:    version.SchemaName,
	}, nil
}
