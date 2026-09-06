package data

import (
	"context"
	"database/sql"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func LoadRegistrationStatusReasons(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	file publication.VerifiedFile,
) (uint64, error) {
	return loadZIPFile(
		ctx,
		connection,
		schemaName,
		"registration_status_reasons",
		[]string{
			"code",
			"name",
		},
		file,
	)
}
