package publication

import (
	"errors"
	"fmt"
)

type VerifyFileProgress func(
	current int,
	total int,
	file ClassifiedFile,
)

func VerifyFiles(
	files []ClassifiedFile,
) ([]VerifiedFile, error) {
	return VerifyFilesWithProgress(files, nil)
}

func VerifyFilesWithProgress(
	files []ClassifiedFile,
	progress VerifyFileProgress,
) ([]VerifiedFile, error) {
	if len(files) == 0 {
		return nil, errors.New(
			"não existem arquivos classificados para verificar",
		)
	}

	result := make([]VerifiedFile, 0, len(files))

	for index, file := range files {
		if progress != nil {
			progress(index+1, len(files), file)
		}

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
