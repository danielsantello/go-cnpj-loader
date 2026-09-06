package data

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"
)

const CurrentSchemaFormatVersion uint16 = 1

var schemaNamePattern = regexp.MustCompile(
	`^[a-z][a-z0-9_]{0,63}$`,
)

func CreateSchema(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	referenceYear uint16,
	referenceMonth uint8,
) error {
	if connection == nil {
		return fmt.Errorf(
			"conexão com o MySQL não foi informada",
		)
	}

	if !schemaNamePattern.MatchString(schemaName) {
		return fmt.Errorf(
			"nome do schema de dados é inválido: %q",
			schemaName,
		)
	}

	createSchemaStatement := fmt.Sprintf(
		`
			CREATE DATABASE %s
			CHARACTER SET utf8mb4
			COLLATE utf8mb4_0900_ai_ci
		`,
		schemaName,
	)

	if _, err := connection.ExecContext(
		ctx,
		createSchemaStatement,
	); err != nil {
		return fmt.Errorf(
			"não foi possível criar o schema de dados %q: %w",
			schemaName,
			err,
		)
	}

	createMetadataStatement := fmt.Sprintf(
		`
			CREATE TABLE %s.schema_metadata (
				id TINYINT UNSIGNED NOT NULL,
				format_version SMALLINT UNSIGNED NOT NULL,
				reference_year SMALLINT UNSIGNED NOT NULL,
				reference_month TINYINT UNSIGNED NOT NULL,
				created_at_utc DATETIME(6) NOT NULL,

				CONSTRAINT pk_schema_metadata
					PRIMARY KEY (id),

				CONSTRAINT chk_schema_metadata_single_row
					CHECK (id = 1),

				CONSTRAINT chk_schema_metadata_reference_year
					CHECK (
						reference_year BETWEEN 1000 AND 9999
					),

				CONSTRAINT chk_schema_metadata_reference_month
					CHECK (
						reference_month BETWEEN 1 AND 12
					)
			)
			ENGINE = InnoDB
			DEFAULT CHARACTER SET = utf8mb4
			COLLATE = utf8mb4_0900_ai_ci
		`,
		schemaName,
	)

	if _, err := connection.ExecContext(
		ctx,
		createMetadataStatement,
	); err != nil {
		return fmt.Errorf(
			"não foi possível criar os metadados do schema %q: %w",
			schemaName,
			err,
		)
	}

	insertMetadataStatement := fmt.Sprintf(
		`
			INSERT INTO %s.schema_metadata (
				id,
				format_version,
				reference_year,
				reference_month,
				created_at_utc
			)
			VALUES (1, ?, ?, ?, ?)
		`,
		schemaName,
	)

	if _, err := connection.ExecContext(
		ctx,
		insertMetadataStatement,
		CurrentSchemaFormatVersion,
		referenceYear,
		referenceMonth,
		time.Now().UTC(),
	); err != nil {
		return fmt.Errorf(
			"não foi possível registrar os metadados do schema %q: %w",
			schemaName,
			err,
		)
	}

	return nil
}
