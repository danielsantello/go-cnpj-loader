package publication

import (
	"errors"
	"fmt"
)

func ValidatePublicationFiles(
	catalog Catalog,
	files []ClassifiedFile,
) error {
	if err := ValidateCatalog(catalog); err != nil {
		return fmt.Errorf(
			"não foi possível validar os arquivos com catálogo inválido: %w",
			err,
		)
	}

	if len(files) == 0 {
		return errors.New(
			"publicação não possui arquivos classificados",
		)
	}

	datasets := make(map[string]Dataset, len(catalog.Datasets))
	partsByDataset := make(
		map[string]map[uint16]struct{},
		len(catalog.Datasets),
	)

	for _, dataset := range catalog.Datasets {
		datasets[dataset.Code] = dataset
	}

	for _, file := range files {
		if _, exists := datasets[file.DatasetCode]; !exists {
			return fmt.Errorf(
				"arquivo %q possui dataset desconhecido: %q",
				file.SourceName,
				file.DatasetCode,
			)
		}

		parts, exists := partsByDataset[file.DatasetCode]
		if !exists {
			parts = make(map[uint16]struct{})
			partsByDataset[file.DatasetCode] = parts
		}

		if _, exists := parts[file.PartNumber]; exists {
			return fmt.Errorf(
				"dataset %q possui a parte %d duplicada",
				file.DatasetCode,
				file.PartNumber,
			)
		}

		parts[file.PartNumber] = struct{}{}
	}

	for _, dataset := range catalog.Datasets {
		parts := partsByDataset[dataset.Code]

		if len(parts) == 0 {
			return fmt.Errorf(
				"publicação não possui arquivos do dataset %q",
				dataset.Code,
			)
		}

		switch dataset.PartNumberRule.Source {
		case PartNumberSourceFixed:
			expectedPart := dataset.PartNumberRule.Value

			if len(parts) != 1 {
				return fmt.Errorf(
					"dataset %q de parte fixa deveria possuir somente a parte %d",
					dataset.Code,
					expectedPart,
				)
			}

			if _, exists := parts[expectedPart]; !exists {
				return fmt.Errorf(
					"dataset %q não possui a parte fixa esperada %d",
					dataset.Code,
					expectedPart,
				)
			}

		case PartNumberSourceCaptureGroup:
			var highestPart uint16

			for partNumber := range parts {
				if partNumber > highestPart {
					highestPart = partNumber
				}
			}

			for expectedPart := uint32(0); expectedPart <= uint32(highestPart); expectedPart++ {
				if _, exists := parts[uint16(expectedPart)]; !exists {
					return fmt.Errorf(
						"dataset %q não possui a parte esperada %d",
						dataset.Code,
						expectedPart,
					)
				}
			}
		}
	}

	return nil
}
