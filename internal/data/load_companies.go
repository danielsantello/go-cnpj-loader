package data

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
	"github.com/go-sql-driver/mysql"
)

func LoadCompaniesFile(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	file publication.VerifiedFile,
) (uint64, error) {
	if connection == nil {
		return 0, fmt.Errorf(
			"conexão com o MySQL não foi informada",
		)
	}

	if !schemaNamePattern.MatchString(schemaName) {
		return 0, fmt.Errorf(
			"nome do schema de dados é inválido: %q",
			schemaName,
		)
	}

	handlerName := fmt.Sprintf(
		"companies_%s_%d",
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
			"não foi possível iniciar a carga de empresas: %w",
			err,
		)
	}
	defer transaction.Rollback()

	statement := fmt.Sprintf(
		`
			LOAD DATA LOCAL INFILE 'Reader::%s'
			INTO TABLE %s.companies
			CHARACTER SET latin1
			FIELDS
				TERMINATED BY ';'
				ENCLOSED BY '"'
				ESCAPED BY ''
			LINES TERMINATED BY '\n'
			(
				@cnpj_root,
				@legal_name,
				@legal_nature_code,
				@responsible_qualification_code,
				@share_capital,
				@company_size_code,
				@responsible_federative_entity
			)
			SET
				cnpj_root = @cnpj_root,
				legal_name = @legal_name,
				legal_nature_code = @legal_nature_code,
				responsible_qualification_code =
					@responsible_qualification_code,
				share_capital = NULLIF(
					REPLACE(
						@share_capital,
						',',
						'.'
					),
					''
				),
				company_size_code = NULLIF(
					@company_size_code,
					''
				),
				responsible_federative_entity = NULLIF(
					@responsible_federative_entity,
					''
				)
		`,
		handlerName,
		schemaName,
	)

	result, err := transaction.ExecContext(ctx, statement)
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível carregar o arquivo %q de empresas: %w",
			file.ClassifiedFile.SourceName,
			err,
		)
	}

	var warningCount uint64

	if err := transaction.QueryRowContext(
		ctx,
		"SHOW COUNT(*) WARNINGS",
	).Scan(&warningCount); err != nil {
		return 0, fmt.Errorf(
			"não foi possível consultar os warnings da carga de empresas: %w",
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
				"carga do arquivo %q de empresas produziu %d warning(s), mas não foi possível consultá-los: %w",
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
					"não foi possível ler os warnings da carga de empresas: %w",
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
				"não foi possível percorrer os warnings da carga de empresas: %w",
				err,
			)
		}

		return 0, fmt.Errorf(
			"carga do arquivo %q de empresas produziu %d warning(s):\n%s",
			file.ClassifiedFile.SourceName,
			warningCount,
			strings.Join(warnings, "\n"),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível obter a quantidade de empresas carregadas: %w",
			err,
		)
	}

	if err := transaction.Commit(); err != nil {
		return 0, fmt.Errorf(
			"não foi possível confirmar a carga do arquivo %q de empresas: %w",
			file.ClassifiedFile.SourceName,
			err,
		)
	}

	return uint64(rowsAffected), nil
}
