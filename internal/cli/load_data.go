package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/danielsantello/go-cnpj-loader/internal/config"
	"github.com/danielsantello/go-cnpj-loader/internal/control"
	"github.com/danielsantello/go-cnpj-loader/internal/data"
	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

type createTableFunction func(
	context.Context,
	*sql.DB,
	string,
) error

type loadTableFunction func(
	context.Context,
	*sql.DB,
	string,
	publication.VerifiedFile,
) (uint64, error)

type referenceTableDefinition struct {
	datasetCode string
	create      createTableFunction
	load        loadTableFunction
}

type dataLoadResult struct {
	economicActivityCount         uint64
	registrationStatusReasonCount uint64
	municipalityCount             uint64
	legalNatureCount              uint64
	countryCount                  uint64
	partnerQualificationCount     uint64
	companyCount                  uint64
	establishmentCount            uint64
	partnerCount                  uint64
	simpleTaxOptionCount          uint64
}

func createVersionAndLoadData(
	ctx context.Context,
	connection *sql.DB,
	value config.Config,
	publicationID uint64,
	referenceYear uint16,
	referenceMonth uint8,
	files []publication.VerifiedFile,
	report func(string),
) (
	resultVersion control.Version,
	resultData dataLoadResult,
	resultErr error,
) {
	version, err := control.CreatePendingVersion(
		ctx,
		connection,
		value.ControlSchema,
		publicationID,
		value.Environment,
	)
	if err != nil {
		return control.Version{}, dataLoadResult{}, fmt.Errorf(
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
		return control.Version{}, dataLoadResult{}, err
	}

	defer func() {
		if resultErr == nil {
			return
		}

		failureContext, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			10*time.Second,
		)
		defer cancel()

		if err := control.MarkVersionFailed(
			failureContext,
			connection,
			value.ControlSchema,
			version.ID,
		); err != nil {
			resultErr = errors.Join(
				resultErr,
				err,
			)
		}
	}()

	if err := data.CreateSchema(
		ctx,
		connection,
		version.SchemaName,
		referenceYear,
		referenceMonth,
	); err != nil {
		return control.Version{}, dataLoadResult{}, err
	}

	definitions := []referenceTableDefinition{
		{
			datasetCode: "economic_activities",
			create:      data.CreateEconomicActivitiesTable,
			load:        data.LoadEconomicActivities,
		},
		{
			datasetCode: "registration_status_reasons",
			create:      data.CreateRegistrationStatusReasonsTable,
			load:        data.LoadRegistrationStatusReasons,
		},
		{
			datasetCode: "municipalities",
			create:      data.CreateMunicipalitiesTable,
			load:        data.LoadMunicipalities,
		},
		{
			datasetCode: "legal_natures",
			create:      data.CreateLegalNaturesTable,
			load:        data.LoadLegalNatures,
		},
		{
			datasetCode: "countries",
			create:      data.CreateCountriesTable,
			load:        data.LoadCountries,
		},
		{
			datasetCode: "partner_qualifications",
			create:      data.CreatePartnerQualificationsTable,
			load:        data.LoadPartnerQualifications,
		},
	}

	rowCounts := make(map[string]uint64, len(definitions))

	for _, definition := range definitions {
		file, err := findVerifiedFile(
			files,
			definition.datasetCode,
		)
		if err != nil {
			return control.Version{}, dataLoadResult{}, err
		}

		if err := definition.create(
			ctx,
			connection,
			version.SchemaName,
		); err != nil {
			return control.Version{}, dataLoadResult{}, err
		}

		rowCount, err := definition.load(
			ctx,
			connection,
			version.SchemaName,
			file,
		)
		if err != nil {
			return control.Version{}, dataLoadResult{}, err
		}

		rowCounts[definition.datasetCode] = rowCount
	}

	companyCount, err := loadCompanies(
		ctx,
		connection,
		version.SchemaName,
		files,
		report,
	)
	if err != nil {
		return control.Version{}, dataLoadResult{}, err
	}

	establishmentCount, err := loadEstablishments(
		ctx,
		connection,
		version.SchemaName,
		files,
		report,
	)
	if err != nil {
		return control.Version{}, dataLoadResult{}, err
	}

	partnerCount, err := loadPartners(
		ctx,
		connection,
		version.SchemaName,
		files,
		report,
	)
	if err != nil {
		return control.Version{}, dataLoadResult{}, err
	}

	simpleTaxOptionCount, err := loadSimpleTaxOptions(
		ctx,
		connection,
		version.SchemaName,
		files,
		report,
	)
	if err != nil {
		return control.Version{}, dataLoadResult{}, err
	}

	if err := control.MarkVersionReady(
		ctx,
		connection,
		value.ControlSchema,
		version.ID,
	); err != nil {
		return control.Version{}, dataLoadResult{}, err
	}

	return version, dataLoadResult{
		economicActivityCount:         rowCounts["economic_activities"],
		registrationStatusReasonCount: rowCounts["registration_status_reasons"],
		municipalityCount:             rowCounts["municipalities"],
		legalNatureCount:              rowCounts["legal_natures"],
		countryCount:                  rowCounts["countries"],
		partnerQualificationCount:     rowCounts["partner_qualifications"],
		companyCount:                  companyCount,
		establishmentCount:            establishmentCount,
		partnerCount:                  partnerCount,
		simpleTaxOptionCount:          simpleTaxOptionCount,
	}, nil
}

func findVerifiedFile(
	files []publication.VerifiedFile,
	datasetCode string,
) (publication.VerifiedFile, error) {
	for _, file := range files {
		if file.ClassifiedFile.DatasetCode == datasetCode {
			return file, nil
		}
	}

	return publication.VerifiedFile{}, fmt.Errorf(
		"publicação não possui o arquivo do dataset %q",
		datasetCode,
	)
}
