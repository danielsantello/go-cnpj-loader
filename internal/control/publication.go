package control

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

var controlSchemaNamePattern = regexp.MustCompile(
	`^[a-z][a-z0-9_]{0,63}$`,
)

type RegisterPublicationInput struct {
	ReferenceYear   uint16
	ReferenceMonth  uint8
	SourceDirectory string
	Files           []publication.VerifiedFile
}

func RegisterAvailablePublication(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	input RegisterPublicationInput,
) (uint64, error) {
	if connection == nil {
		return 0, fmt.Errorf(
			"conexão com o MySQL não foi informada",
		)
	}

	if !controlSchemaNamePattern.MatchString(schemaName) {
		return 0, fmt.Errorf(
			"nome do schema de controle é inválido: %q",
			schemaName,
		)
	}

	if input.ReferenceYear < 1000 || input.ReferenceYear > 9999 {
		return 0, fmt.Errorf(
			"ano de referência é inválido: %d",
			input.ReferenceYear,
		)
	}

	if input.ReferenceMonth < 1 || input.ReferenceMonth > 12 {
		return 0, fmt.Errorf(
			"mês de referência é inválido: %d",
			input.ReferenceMonth,
		)
	}

	if strings.TrimSpace(input.SourceDirectory) == "" {
		return 0, fmt.Errorf(
			"diretório da publicação é obrigatório",
		)
	}

	if !filepath.IsAbs(input.SourceDirectory) {
		return 0, fmt.Errorf(
			"diretório da publicação deve ser absoluto",
		)
	}

	contentFingerprint, err := publication.CalculateContentFingerprint(
		input.Files,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível calcular o fingerprint da publicação: %w",
			err,
		)
	}

	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível iniciar o registro da publicação: %w",
			err,
		)
	}
	defer transaction.Rollback()

	registeredAt := time.Now().UTC()

	publicationStatement := fmt.Sprintf(
		`
			INSERT INTO %s.publications (
				reference_year,
				reference_month,
				source_location,
				content_fingerprint,
				registered_at_utc
			)
			VALUES (?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
					id = LAST_INSERT_ID(id)
		`,
		schemaName,
	)

	result, err := transaction.ExecContext(
		ctx,
		publicationStatement,
		input.ReferenceYear,
		input.ReferenceMonth,
		input.SourceDirectory,
		contentFingerprint[:],
		registeredAt,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível registrar a publicação: %w",
			err,
		)
	}

	publicationID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf(
			"não foi possível obter o identificador da publicação: %w",
			err,
		)
	}

	fileStatement := fmt.Sprintf(
		`
			INSERT INTO %s.publication_files (
				publication_id,
				dataset_code,
				part_number,
				source_name,
				source_location,
				size_bytes,
				sha256,
				registered_at_utc
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
					id = id
		`,
		schemaName,
	)

	for _, file := range input.Files {
		classifiedFile := file.ClassifiedFile

		if _, err := transaction.ExecContext(
			ctx,
			fileStatement,
			publicationID,
			classifiedFile.DatasetCode,
			classifiedFile.PartNumber,
			classifiedFile.SourceName,
			classifiedFile.SourceLocation,
			file.SizeBytes,
			file.SHA256[:],
			registeredAt,
		); err != nil {
			return 0, fmt.Errorf(
				"não foi possível registrar o arquivo %q: %w",
				classifiedFile.SourceName,
				err,
			)
		}
	}

	if err := transaction.Commit(); err != nil {
		return 0, fmt.Errorf(
			"não foi possível concluir o registro da publicação: %w",
			err,
		)
	}

	return uint64(publicationID), nil
}
