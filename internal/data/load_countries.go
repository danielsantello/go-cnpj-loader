package data

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
	"github.com/go-sql-driver/mysql"
)

type errorReader struct {
	err error
}

func (reader errorReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func LoadCountries(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	file publication.VerifiedFile,
) (uint64, error) {
	if connection == nil {
		return 0, fmt.Errorf("conexão com o MySQL não foi informada")
	}

	if !schemaNamePattern.MatchString(schemaName) {
		return 0, fmt.Errorf(
			"nome do schema de dados é inválido: %q",
			schemaName,
		)
	}

	handlerName := "countries_" + schemaName

	mysql.RegisterReaderHandler(
		handlerName,
		func() io.Reader {
			reader, err := publication.OpenZIPContent(file)
			if err != nil {
				return errorReader{err: err}
			}

			return reader
		},
	)
	defer mysql.DeregisterReaderHandler(handlerName)

	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível iniciar a transação da carga de países: %w",
			err,
		)
	}
	defer transaction.Rollback()

	statement := fmt.Sprintf(
		`
			LOAD DATA LOCAL INFILE 'Reader::%s'
			INTO TABLE %s.countries
			CHARACTER SET latin1
			FIELDS
				TERMINATED BY ';'
				ENCLOSED BY '"'
			LINES TERMINATED BY '\n'
			(
				code,
				name
			)
		`,
		handlerName,
		schemaName,
	)

	result, err := transaction.ExecContext(ctx, statement)
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível carregar o arquivo de países: %w",
			err,
		)
	}

	var warningCount uint64

	if err := transaction.QueryRowContext(
		ctx,
		"SHOW COUNT(*) WARNINGS",
	).Scan(&warningCount); err != nil {
		return 0, fmt.Errorf(
			"não foi possível consultar os warnings da carga de países: %w",
			err,
		)
	}

	if warningCount != 0 {
		return 0, fmt.Errorf(
			"carga de países produziu %d warning(s)",
			warningCount,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível obter a quantidade de países carregados: %w",
			err,
		)
	}

	if err := transaction.Commit(); err != nil {
		return 0, fmt.Errorf(
			"não foi possível confirmar a carga de países: %w",
			err,
		)
	}

	return uint64(rowsAffected), nil
}
