package cli

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-loader/internal/config"
	"github.com/danielsantello/go-cnpj-loader/internal/control"
	"github.com/danielsantello/go-cnpj-loader/internal/data"
	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func createVersionAndLoadCountries(
	ctx context.Context,
	connection *sql.DB,
	value config.Config,
	publicationID uint64,
	referenceYear uint16,
	referenceMonth uint8,
	files []publication.VerifiedFile,
) (control.Version, uint64, error) {
	version, err := control.CreatePendingVersion(
		ctx,
		connection,
		value.ControlSchema,
		publicationID,
		value.Environment,
	)
	if err != nil {
		return control.Version{}, 0, fmt.Errorf(
			"não foi possível criar a versão: %w",
			err,
		)
	}

	if err := control.MarkVersionLoading(
		ctx,
		connection,
		value.ControlSchema,
		version.ID,
	); err != nil {
		return control.Version{}, 0, err
	}

	if err := data.CreateSchema(
		ctx,
		connection,
		version.SchemaName,
		referenceYear,
		referenceMonth,
	); err != nil {
		return control.Version{}, 0, err
	}

	if err := data.CreateCountriesTable(
		ctx,
		connection,
		version.SchemaName,
	); err != nil {
		return control.Version{}, 0, err
	}

	var countriesFile publication.VerifiedFile
	foundCountries := false

	for _, file := range files {
		if file.ClassifiedFile.DatasetCode == "countries" {
			countriesFile = file
			foundCountries = true
			break
		}
	}

	if !foundCountries {
		return control.Version{}, 0, fmt.Errorf(
			"publicação não possui o arquivo do dataset %q",
			"countries",
		)
	}

	rowCount, err := data.LoadCountries(
		ctx,
		connection,
		version.SchemaName,
		countriesFile,
	)
	if err != nil {
		return control.Version{}, 0, err
	}

	return version, rowCount, nil
}
