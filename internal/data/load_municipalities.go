package data

import (
	"context"
	"database/sql"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func LoadMunicipalities(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	file publication.VerifiedFile,
) (uint64, error) {
	return loadZIPFile(
		ctx,
		connection,
		schemaName,
		"municipalities",
		[]string{
			"code",
			"name",
		},
		file,
	)
}
