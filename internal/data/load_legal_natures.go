package data

import (
	"context"
	"database/sql"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func LoadLegalNatures(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	file publication.VerifiedFile,
) (uint64, error) {
	return loadZIPFile(
		ctx,
		connection,
		schemaName,
		"legal_natures",
		[]string{
			"code",
			"name",
		},
		file,
	)
}
