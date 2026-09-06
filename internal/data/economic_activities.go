package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func CreateEconomicActivitiesTable(
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
			CREATE TABLE %s.economic_activities (
				code CHAR(7) NOT NULL,
				name VARCHAR(255) NOT NULL
			)
			ENGINE = InnoDB
			DEFAULT CHARACTER SET = utf8mb4
			COLLATE = utf8mb4_0900_ai_ci
		`,
		schemaName,
	)

	if _, err := connection.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf(
			"não foi possível criar a tabela de atividades econômicas no schema %q: %w",
			schemaName,
			err,
		)
	}

	return nil
}
