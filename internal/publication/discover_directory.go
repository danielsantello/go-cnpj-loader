package publication

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DiscoverDirectory(sourceDirectory string) ([]DiscoveredFile, error) {
	if strings.TrimSpace(sourceDirectory) == "" {
		return nil, errors.New(
			"localização do diretório da publicação é obrigatória",
		)
	}

	if !filepath.IsAbs(sourceDirectory) {
		return nil, errors.New(
			"localização do diretório da publicação deve ser absoluta",
		)
	}

	entries, err := os.ReadDir(sourceDirectory)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível ler o diretório da publicação %q: %w",
			sourceDirectory,
			err,
		)
	}

	result := make([]DiscoveredFile, 0)

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf(
				"não foi possível consultar o arquivo %q: %w",
				entry.Name(),
				err,
			)
		}

		if !info.Mode().IsRegular() {
			continue
		}

		if !strings.EqualFold(
			filepath.Ext(entry.Name()),
			".zip",
		) {
			continue
		}

		result = append(
			result,
			DiscoveredFile{
				SourceName: entry.Name(),
				SourceLocation: filepath.Join(
					sourceDirectory,
					entry.Name(),
				),
			},
		)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf(
			"diretório da publicação %q não possui arquivos ZIP",
			sourceDirectory,
		)
	}

	return result, nil
}
