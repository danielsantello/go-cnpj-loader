package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func CreateEstablishmentsTable(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
) error {
	if connection == nil {
		return errors.New(
			"conexão com o MySQL não foi informada",
		)
	}

	if !schemaNamePattern.MatchString(schemaName) {
		return fmt.Errorf(
			"nome do schema de dados é inválido: %q",
			schemaName,
		)
	}

	statement := fmt.Sprintf(
		`
			CREATE TABLE %s.establishments (
				cnpj CHAR(14) NOT NULL,
				cnpj_root CHAR(8) NOT NULL,
				branch_number CHAR(4) NOT NULL,
				check_digits CHAR(2) NOT NULL,
				head_office_branch_indicator CHAR(1) NOT NULL,
				trade_name VARCHAR(255) NULL,
				registration_status_code CHAR(2) NULL,
				registration_status_date DATE NULL,
				registration_status_reason_code CHAR(2) NULL,
				foreign_city_name VARCHAR(150) NULL,
				country_code CHAR(3) NULL,
				activity_start_date DATE NULL,
				main_economic_activity_code CHAR(7) NULL,
				secondary_economic_activities VARCHAR(4000) NULL,
				street_type VARCHAR(50) NULL,
				street_name VARCHAR(255) NULL,
				street_number VARCHAR(20) NULL,
				address_complement VARCHAR(255) NULL,
				neighborhood VARCHAR(100) NULL,
				postal_code CHAR(8) NULL,
				state_code CHAR(2) NULL,
				municipality_code CHAR(4) NULL,
				phone_area_code_1 VARCHAR(4) NULL,
				phone_number_1 VARCHAR(12) NULL,
				phone_area_code_2 VARCHAR(4) NULL,
				phone_number_2 VARCHAR(12) NULL,
				fax_area_code VARCHAR(4) NULL,
				fax_number VARCHAR(12) NULL,
				email_address VARCHAR(255) NULL,
				special_status VARCHAR(50) NULL,
				special_status_date DATE NULL
			)
			ENGINE = InnoDB
			DEFAULT CHARACTER SET = utf8mb4
			COLLATE = utf8mb4_0900_ai_ci
		`,
		schemaName,
	)

	if _, err := connection.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf(
			"não foi possível criar a tabela de estabelecimentos no schema %q: %w",
			schemaName,
			err,
		)
	}

	return nil
}
