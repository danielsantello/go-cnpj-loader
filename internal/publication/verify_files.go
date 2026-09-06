package publication

import (
	"errors"
	"fmt"
)

func VerifyFiles(
	files []ClassifiedFile,
) ([]VerifiedFile, error) {
	if len(files) == 0 {
		return nil, errors.New(
			"não existem arquivos classificados para verificar",
		)
	}

	result := make([]VerifiedFile, 0, len(files))

	for _, file := range files {
		verifiedFile, err := VerifyFile(file)
		if err != nil {
			return nil, fmt.Errorf(
				"não foi possível verificar os arquivos da publicação: %w",
				err,
			)
		}

		result = append(result, verifiedFile)
	}

	return result, nil
}
