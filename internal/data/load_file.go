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

type zipContentErrorReader struct {
	err error
}

func (reader zipContentErrorReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func loadZIPFile(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	tableName string,
	columns []string,
	file publication.VerifiedFile,
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

	if len(columns) == 0 {
		return 0, errors.New(
			"nenhuma coluna foi informada para a carga",
		)
	}

	for _, column := range columns {
		if !schemaNamePattern.MatchString(column) {
			return 0, fmt.Errorf(
				"nome da coluna de dados é inválido: %q",
				column,
			)
		}
	}

	handlerName := tableName + "_" + schemaName

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

	statement := fmt.Sprintf(
		`
			LOAD DATA LOCAL INFILE 'Reader::%s'
			INTO TABLE %s.%s
			CHARACTER SET latin1
			FIELDS
				TERMINATED BY ';'
				ENCLOSED BY '"'
			LINES TERMINATED BY '\n'
			(
				%s
			)
		`,
		handlerName,
		schemaName,
		tableName,
		strings.Join(columns, ",\n"),
	)

	result, err := transaction.ExecContext(ctx, statement)
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível carregar a tabela %q: %w",
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
		return 0, fmt.Errorf(
			"carga da tabela %q produziu %d warning(s)",
			tableName,
			warningCount,
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
			"não foi possível confirmar a carga da tabela %q: %w",
			tableName,
			err,
		)
	}

	return uint64(rowsAffected), nil
}
