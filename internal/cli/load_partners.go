package cli

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/danielsantello/go-cnpj-loader/internal/data"
	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func loadPartners(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	files []publication.VerifiedFile,
	report func(string),
) (uint64, error) {
	if err := data.CreatePartnersTable(
		ctx,
		connection,
		schemaName,
	); err != nil {
		return 0, err
	}

	var totalRows uint64
	foundFiles := 0

	for _, file := range files {
		if file.ClassifiedFile.DatasetCode != "partners" {
			continue
		}

		foundFiles++

		if report != nil {
			report(fmt.Sprintf(
				"Carregando %s...",
				file.ClassifiedFile.SourceName,
			))
		}

		startedAt := time.Now()

		rowCount, err := data.LoadPartnersFile(
			ctx,
			connection,
			schemaName,
			file,
		)
		if err != nil {
			return 0, err
		}

		duration := time.Since(startedAt).Round(time.Millisecond)
		totalRows += rowCount

		if report != nil {
			report(fmt.Sprintf(
				"%s: %d registros carregados em %s.",
				file.ClassifiedFile.SourceName,
				rowCount,
				duration,
			))
		}
	}

	if foundFiles == 0 {
		return 0, fmt.Errorf(
			"publicação não possui arquivos do dataset %q",
			"partners",
		)
	}

	return totalRows, nil
}
