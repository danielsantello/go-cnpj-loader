package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func CreatePartnersTable(
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
			CREATE TABLE %s.partners (
				cnpj_root CHAR(8) NOT NULL,
				partner_type_code CHAR(1) NOT NULL,
				partner_name VARCHAR(255) NOT NULL,
				partner_document VARCHAR(20) NULL,
				partner_qualification_code CHAR(2) NULL,
				entry_date DATE NULL,
				country_code CHAR(3) NULL,
				legal_representative_document VARCHAR(20) NULL,
				legal_representative_name VARCHAR(255) NULL,
				legal_representative_qualification_code CHAR(2) NULL,
				age_range_code CHAR(1) NULL
			)
			ENGINE = InnoDB
			DEFAULT CHARACTER SET = utf8mb4
			COLLATE = utf8mb4_0900_ai_ci
		`,
		schemaName,
	)

	if _, err := connection.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf(
			"não foi possível criar a tabela de sócios no schema %q: %w",
			schemaName,
			err,
		)
	}

	return nil
}
