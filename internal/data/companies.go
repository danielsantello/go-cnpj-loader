package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func CreateCompaniesTable(
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
			CREATE TABLE %s.companies (
				cnpj_root CHAR(8) NOT NULL,
				legal_name VARCHAR(255) NOT NULL,
				legal_nature_code CHAR(4) NOT NULL,
				responsible_qualification_code CHAR(2) NOT NULL,
				share_capital DECIMAL(20, 2) NULL,
				company_size_code CHAR(2) NULL,
				responsible_federative_entity VARCHAR(100) NULL
			)
			ENGINE = InnoDB
			DEFAULT CHARACTER SET = utf8mb4
			COLLATE = utf8mb4_0900_ai_ci
		`,
		schemaName,
	)

	if _, err := connection.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf(
			"não foi possível criar a tabela de empresas no schema %q: %w",
			schemaName,
			err,
		)
	}

	return nil
}
