package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/danielsantello/go-cnpj-loader/internal/buildinfo"
	"github.com/danielsantello/go-cnpj-loader/internal/config"
	"github.com/danielsantello/go-cnpj-loader/internal/control/migrations"
	"github.com/danielsantello/go-cnpj-loader/internal/database"
	"github.com/spf13/cobra"
)

func newMigrateControlCommand(info buildinfo.Info) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate-control",
		Short: "Cria ou atualiza o schema de controle",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			schemaName, err := migrateControl(
				command.Context(),
				info,
			)
			if err != nil {
				return err
			}

			fmt.Fprintf(
				command.OutOrStdout(),
				"Schema de controle %q atualizado com sucesso.\n",
				schemaName,
			)

			return nil
		},
	}
}

func migrateControl(
	ctx context.Context,
	info buildinfo.Info,
) (_ string, resultErr error) {
	value, err := config.Load()
	if err != nil {
		return "", fmt.Errorf(
			"não foi possível carregar a configuração: %w",
			err,
		)
	}

	if err := config.Validate(value); err != nil {
		return "", fmt.Errorf(
			"configuração inválida: %w",
			err,
		)
	}

	connection, err := database.OpenMySQL(value.MySQL)
	if err != nil {
		return "", fmt.Errorf(
			"não foi possível abrir o gerenciador de conexões: %w",
			err,
		)
	}
	defer func() {
		if err := connection.Close(); err != nil {
			resultErr = errors.Join(
				resultErr,
				fmt.Errorf(
					"não foi possível fechar a conexão com o MySQL: %w",
					err,
				),
			)
		}
	}()

	pingContext, cancel := context.WithTimeout(
		ctx,
		value.MySQL.ConnectTimeout,
	)
	defer cancel()

	if err := database.PingMySQL(
		pingContext,
		connection,
	); err != nil {
		return "", err
	}

	if err := migrations.Migrate(
		ctx,
		connection,
		value.ControlSchema,
		info,
	); err != nil {
		return "", fmt.Errorf(
			"não foi possível migrar o schema de controle: %w",
			err,
		)
	}

	return value.ControlSchema, nil
}
