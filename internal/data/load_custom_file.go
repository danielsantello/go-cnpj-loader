package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
	"github.com/go-sql-driver/mysql"
)

type loadStatementBuilder func(handlerName string) string

func loadCustomZIPFile(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	tableName string,
	file publication.VerifiedFile,
	buildStatement loadStatementBuilder,
) (uint64, error) {
	if connection == nil {
		return 0, errors.New(
			"conexão com o MySQL não foi informada",
		)
	}

	if !schemaNamePattern.MatchString(schemaName) {
		return 0, fmt.Errorf(
			"nome do schema de dados é inválido: %q",
			schemaName,
		)
	}

	if !schemaNamePattern.MatchString(tableName) {
		return 0, fmt.Errorf(
			"nome da tabela de dados é inválido: %q",
			tableName,
		)
	}

	if buildStatement == nil {
		return 0, errors.New(
			"construtor do comando de carga não foi informado",
		)
	}

	handlerName := fmt.Sprintf(
		"%s_%s_%d",
		tableName,
		schemaName,
		file.ClassifiedFile.PartNumber,
	)

	mysql.RegisterReaderHandler(
		handlerName,
		func() io.Reader {
			reader, err := publication.OpenZIPContent(file)
			if err != nil {
				return zipContentErrorReader{err: err}
			}

			return reader
		},
	)
	defer mysql.DeregisterReaderHandler(handlerName)

	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível iniciar a carga da tabela %q: %w",
			tableName,
			err,
		)
	}
	defer transaction.Rollback()

	result, err := transaction.ExecContext(
		ctx,
		buildStatement(handlerName),
	)
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível carregar o arquivo %q na tabela %q: %w",
			file.ClassifiedFile.SourceName,
			tableName,
			err,
		)
	}

	var warningCount uint64

	if err := transaction.QueryRowContext(
		ctx,
		"SHOW COUNT(*) WARNINGS",
	).Scan(&warningCount); err != nil {
		return 0, fmt.Errorf(
			"não foi possível consultar os warnings da tabela %q: %w",
			tableName,
			err,
		)
	}

	if warningCount != 0 {
		rows, err := transaction.QueryContext(
			ctx,
			"SHOW WARNINGS LIMIT 10",
		)
		if err != nil {
			return 0, fmt.Errorf(
				"carga do arquivo %q produziu %d warning(s), mas não foi possível consultá-los: %w",
				file.ClassifiedFile.SourceName,
				warningCount,
				err,
			)
		}
		defer rows.Close()

		warnings := make([]string, 0, warningCount)

		for rows.Next() {
			var (
				level   string
				code    uint16
				message string
			)

			if err := rows.Scan(
				&level,
				&code,
				&message,
			); err != nil {
				return 0, fmt.Errorf(
					"não foi possível ler os warnings da tabela %q: %w",
					tableName,
					err,
				)
			}

			warnings = append(
				warnings,
				fmt.Sprintf(
					"%s %d: %s",
					level,
					code,
					message,
				),
			)
		}

		if err := rows.Err(); err != nil {
			return 0, fmt.Errorf(
				"não foi possível percorrer os warnings da tabela %q: %w",
				tableName,
				err,
			)
		}

		return 0, fmt.Errorf(
			"carga do arquivo %q na tabela %q produziu %d warning(s):\n%s",
			file.ClassifiedFile.SourceName,
			tableName,
			warningCount,
			strings.Join(warnings, "\n"),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível obter a quantidade de registros da tabela %q: %w",
			tableName,
			err,
		)
	}

	if err := transaction.Commit(); err != nil {
		return 0, fmt.Errorf(
			"não foi possível confirmar a carga do arquivo %q na tabela %q: %w",
			file.ClassifiedFile.SourceName,
			tableName,
			err,
		)
	}

	return uint64(rowsAffected), nil
}
