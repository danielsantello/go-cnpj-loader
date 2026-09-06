package publication

import "fmt"

func DiscoverDirectoryPublication(
	catalog Catalog,
	source Source,
) ([]VerifiedFile, error) {
	return DiscoverDirectoryPublicationWithProgress(
		catalog,
		source,
		nil,
	)
}

func DiscoverDirectoryPublicationWithProgress(
	catalog Catalog,
	source Source,
	progress VerifyFileProgress,
) ([]VerifiedFile, error) {
	discoveredFiles, err := DiscoverDirectory(source)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível descobrir a publicação no diretório: %w",
			err,
		)
	}

	classifiedFiles, err := ClassifyFiles(
		catalog,
		discoveredFiles,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível classificar a publicação descoberta: %w",
			err,
		)
	}

	if err := ValidatePublicationFiles(
		catalog,
		classifiedFiles,
	); err != nil {
		return nil, fmt.Errorf(
			"publicação descoberta no diretório é inválida: %w",
			err,
		)
	}

	verifiedFiles, err := VerifyFilesWithProgress(
		classifiedFiles,
		progress,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível verificar a publicação descoberta: %w",
			err,
		)
	}

	return verifiedFiles, nil
}
