package cli

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/danielsantello/go-cnpj-loader/internal/data"
	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func loadSimpleTaxOptions(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	files []publication.VerifiedFile,
	report func(string),
) (uint64, error) {
	file, err := findVerifiedFile(
		files,
		"simple_tax_options",
	)
	if err != nil {
		return 0, err
	}

	if err := data.CreateSimpleTaxOptionsTable(
		ctx,
		connection,
		schemaName,
	); err != nil {
		return 0, err
	}

	if report != nil {
		report(fmt.Sprintf(
			"Carregando %s...",
			file.ClassifiedFile.SourceName,
		))
	}

	startedAt := time.Now()

	rowCount, err := data.LoadSimpleTaxOptions(
		ctx,
		connection,
		schemaName,
		file,
	)
	if err != nil {
		return 0, err
	}

	duration := time.Since(startedAt).Round(time.Millisecond)

	if report != nil {
		report(fmt.Sprintf(
			"%s: %d registros carregados em %s.",
			file.ClassifiedFile.SourceName,
			rowCount,
			duration,
		))
	}

	return rowCount, nil
}
