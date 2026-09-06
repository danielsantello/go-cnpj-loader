package publication

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func ClassifyFile(
	catalog Catalog,
	file DiscoveredFile,
) (ClassifiedFile, error) {
	if err := ValidateCatalog(catalog); err != nil {
		return ClassifiedFile{}, fmt.Errorf(
			"não foi possível classificar o arquivo com catálogo inválido: %w",
			err,
		)
	}

	if strings.TrimSpace(file.SourceName) == "" {
		return ClassifiedFile{}, errors.New(
			"nome do arquivo descoberto é obrigatório",
		)
	}

	if strings.TrimSpace(file.SourceLocation) == "" {
		return ClassifiedFile{}, errors.New(
			"localização do arquivo descoberto é obrigatória",
		)
	}

	var result ClassifiedFile
	found := false

	for _, dataset := range catalog.Datasets {
		pattern := regexp.MustCompile(dataset.FilePattern)
		matches := pattern.FindStringSubmatch(file.SourceName)

		if matches == nil {
			continue
		}

		if found {
			return ClassifiedFile{}, fmt.Errorf(
				"arquivo %q corresponde a mais de um dataset",
				file.SourceName,
			)
		}

		partNumber := dataset.PartNumberRule.Value

		if dataset.PartNumberRule.Source ==
			PartNumberSourceCaptureGroup {
			captureGroup := dataset.PartNumberRule.CaptureGroup
			capturedValue := matches[captureGroup]

			parsedPartNumber, err := strconv.ParseUint(
				capturedValue,
				10,
				16,
			)
			if err != nil {
				return ClassifiedFile{}, fmt.Errorf(
					"arquivo %q possui número de parte inválido %q: %w",
					file.SourceName,
					capturedValue,
					err,
				)
			}

			partNumber = uint16(parsedPartNumber)
		}

		result = ClassifiedFile{
			DatasetCode:    dataset.Code,
			PartNumber:     partNumber,
			SourceName:     file.SourceName,
			SourceLocation: file.SourceLocation,
		}
		found = true
	}

	if !found {
		return ClassifiedFile{}, fmt.Errorf(
			"arquivo %q não corresponde a nenhum dataset conhecido",
			file.SourceName,
		)
	}

	return result, nil
}
