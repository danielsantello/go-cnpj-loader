package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func CreateSimpleTaxOptionsTable(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
) error {
	if connection == nil {
		return errors.New("conexão com o MySQL não foi informada")
	}

	if !schemaNamePattern.MatchString(schemaName) {
		return fmt.Errorf(
			"nome do schema de dados é inválido: %q",
			schemaName,
		)
	}

	statement := fmt.Sprintf(
		`
			CREATE TABLE %s.simple_tax_options (
				cnpj_root CHAR(8) NOT NULL,
				simple_option_indicator CHAR(1) NULL,
				simple_option_date DATE NULL,
				simple_exclusion_date DATE NULL,
				mei_option_indicator CHAR(1) NULL,
				mei_option_date DATE NULL,
				mei_exclusion_date DATE NULL
			)
			ENGINE = InnoDB
			DEFAULT CHARACTER SET = utf8mb4
			COLLATE = utf8mb4_0900_ai_ci
		`,
		schemaName,
	)

	if _, err := connection.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf(
			"não foi possível criar a tabela de opções tributárias no schema %q: %w",
			schemaName,
			err,
		)
	}

	return nil
}
