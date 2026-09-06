package publication

import (
	"errors"
	"fmt"
	"sort"
)

type datasetPart struct {
	datasetCode string
	partNumber  uint16
}

func ClassifyFiles(
	catalog Catalog,
	files []DiscoveredFile,
) ([]ClassifiedFile, error) {
	if len(files) == 0 {
		return nil, errors.New(
			"não existem arquivos descobertos para classificar",
		)
	}

	datasetOrder := make(map[string]int, len(catalog.Datasets))

	for index, dataset := range catalog.Datasets {
		datasetOrder[dataset.Code] = index
	}

	result := make([]ClassifiedFile, 0, len(files))
	knownParts := make(map[datasetPart]string)

	for _, file := range files {
		classifiedFile, err := ClassifyFile(catalog, file)
		if err != nil {
			return nil, fmt.Errorf(
				"não foi possível classificar os arquivos da publicação: %w",
				err,
			)
		}

		key := datasetPart{
			datasetCode: classifiedFile.DatasetCode,
			partNumber:  classifiedFile.PartNumber,
		}

		if previousName, exists := knownParts[key]; exists {
			return nil, fmt.Errorf(
				"arquivos %q e %q representam o mesmo dataset %q, parte %d",
				previousName,
				classifiedFile.SourceName,
				classifiedFile.DatasetCode,
				classifiedFile.PartNumber,
			)
		}

		knownParts[key] = classifiedFile.SourceName
		result = append(result, classifiedFile)
	}

	sort.Slice(
		result,
		func(left int, right int) bool {
			leftFile := result[left]
			rightFile := result[right]

			leftOrder := datasetOrder[leftFile.DatasetCode]
			rightOrder := datasetOrder[rightFile.DatasetCode]

			if leftOrder != rightOrder {
				return leftOrder < rightOrder
			}

			return leftFile.PartNumber < rightFile.PartNumber
		},
	)

	return result, nil
}
