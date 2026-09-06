package control

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/danielsantello/go-cnpj-loader/internal/config"
)

type Version struct {
	ID             uint64
	PublicationID  uint64
	SchemaName     string
	SequenceNumber uint32
}

func CreatePendingVersion(
	ctx context.Context,
	connection *sql.DB,
	controlSchema string,
	publicationID uint64,
	environment config.Environment,
) (Version, error) {
	if connection == nil {
		return Version{}, fmt.Errorf(
			"conexão com o MySQL não foi informada",
		)
	}

	if !controlSchemaNamePattern.MatchString(controlSchema) {
		return Version{}, fmt.Errorf(
			"nome do schema de controle é inválido: %q",
			controlSchema,
		)
	}

	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return Version{}, fmt.Errorf(
			"não foi possível iniciar a criação da versão: %w",
			err,
		)
	}
	defer transaction.Rollback()

	publicationQuery := fmt.Sprintf(
		`
			SELECT
				reference_year,
				reference_month
			FROM %s.publications
			WHERE id = ?
				AND status = 'available'
			FOR UPDATE
		`,
		controlSchema,
	)

	var (
		referenceYear  uint16
		referenceMonth uint8
	)

	if err := transaction.QueryRowContext(
		ctx,
		publicationQuery,
		publicationID,
	).Scan(
		&referenceYear,
		&referenceMonth,
	); err != nil {
		return Version{}, fmt.Errorf(
			"não foi possível localizar a publicação disponível %d: %w",
			publicationID,
			err,
		)
	}

	sequenceQuery := fmt.Sprintf(
		`
			SELECT COALESCE(MAX(sequence_number), 0) + 1
			FROM %s.versions
			WHERE publication_id = ?
				AND environment = ?
		`,
		controlSchema,
	)

	var sequenceNumber uint32

	if err := transaction.QueryRowContext(
		ctx,
		sequenceQuery,
		publicationID,
		environment,
	).Scan(&sequenceNumber); err != nil {
		return Version{}, fmt.Errorf(
			"não foi possível definir a sequência da versão: %w",
			err,
		)
	}

	schemaName := fmt.Sprintf(
		"cnpj_%04d_%02d_%03d",
		referenceYear,
		referenceMonth,
		sequenceNumber,
	)

	createdAt := time.Now().UTC()

	insertStatement := fmt.Sprintf(
		`
			INSERT INTO %s.versions (
				publication_id,
				schema_name,
				environment,
				sequence_number,
				status,
				created_at_utc,
				status_changed_at_utc
			)
			VALUES (?, ?, ?, ?, 'pending', ?, ?)
		`,
		controlSchema,
	)

	result, err := transaction.ExecContext(
		ctx,
		insertStatement,
		publicationID,
		schemaName,
		environment,
		sequenceNumber,
		createdAt,
		createdAt,
	)
	if err != nil {
		return Version{}, fmt.Errorf(
			"não foi possível registrar a versão pendente: %w",
			err,
		)
	}

	versionID, err := result.LastInsertId()
	if err != nil {
		return Version{}, fmt.Errorf(
			"não foi possível obter o identificador da versão: %w",
			err,
		)
	}

	if err := transaction.Commit(); err != nil {
		return Version{}, fmt.Errorf(
			"não foi possível concluir a criação da versão: %w",
			err,
		)
	}

	return Version{
		ID:             uint64(versionID),
		PublicationID:  publicationID,
		SchemaName:     schemaName,
		SequenceNumber: sequenceNumber,
	}, nil
}

func MarkVersionLoading(
	ctx context.Context,
	connection *sql.DB,
	controlSchema string,
	versionID uint64,
) error {
	if connection == nil {
		return fmt.Errorf(
			"conexão com o MySQL não foi informada",
		)
	}

	if !controlSchemaNamePattern.MatchString(controlSchema) {
		return fmt.Errorf(
			"nome do schema de controle é inválido: %q",
			controlSchema,
		)
	}

	statement := fmt.Sprintf(
		`
			UPDATE %s.versions
			SET
				status = 'loading',
				status_changed_at_utc = ?
			WHERE id = ?
				AND status = 'pending'
		`,
		controlSchema,
	)

	result, err := connection.ExecContext(
		ctx,
		statement,
		time.Now().UTC(),
		versionID,
	)
	if err != nil {
		return fmt.Errorf(
			"não foi possível iniciar a carga da versão %d: %w",
			versionID,
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"não foi possível confirmar a atualização da versão %d: %w",
			versionID,
			err,
		)
	}

	if rowsAffected != 1 {
		return fmt.Errorf(
			"versão %d não está pendente para iniciar a carga",
			versionID,
		)
	}

	return nil
}
