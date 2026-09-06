package publication

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DiscoverDirectory(source Source) ([]DiscoveredFile, error) {
	if source.Type != SourceTypeDirectory {
		return nil, fmt.Errorf(
			"origem deveria ser do tipo %q, mas recebeu %q",
			SourceTypeDirectory,
			source.Type,
		)
	}

	if strings.TrimSpace(source.Location) == "" {
		return nil, errors.New(
			"localização do diretório da publicação é obrigatória",
		)
	}

	if !filepath.IsAbs(source.Location) {
		return nil, errors.New(
			"localização do diretório da publicação deve ser absoluta",
		)
	}

	entries, err := os.ReadDir(source.Location)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível ler o diretório da publicação %q: %w",
			source.Location,
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
					source.Location,
					entry.Name(),
				),
			},
		)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf(
			"diretório da publicação %q não possui arquivos ZIP",
			source.Location,
		)
	}

	return result, nil
}
