package migrations

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
)

const sqlDirectory = "sql"

//go:embed sql/*.sql
var sqlFiles embed.FS

type Migration struct {
	Version  uint32
	Name     string
	FileName string
	SQL      string
	Checksum [sha256.Size]byte
}

func LoadCatalog() ([]Migration, error) {
	entries, err := fs.ReadDir(sqlFiles, sqlDirectory)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível ler os arquivos de migrations: %w",
			err,
		)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("nenhuma migration foi encontrada")
	}

	result := make([]Migration, 0, len(entries))
	versions := make(map[uint32]string, len(entries))
	names := make(map[string]string, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		version, name, err := parseFileName(entry.Name())
		if err != nil {
			return nil, err
		}

		if previousFile, exists := versions[version]; exists {
			return nil, fmt.Errorf(
				"versão %d está duplicada nos arquivos %q e %q",
				version,
				previousFile,
				entry.Name(),
			)
		}

		if previousFile, exists := names[name]; exists {
			return nil, fmt.Errorf(
				"nome %q está duplicado nos arquivos %q e %q",
				name,
				previousFile,
				entry.Name(),
			)
		}

		filePath := path.Join(sqlDirectory, entry.Name())
		content, err := sqlFiles.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf(
				"não foi possível ler a migration %q: %w",
				entry.Name(),
				err,
			)
		}

		if strings.TrimSpace(string(content)) == "" {
			return nil, fmt.Errorf(
				"migration %q está vazia",
				entry.Name(),
			)
		}

		result = append(result, Migration{
			Version:  version,
			Name:     name,
			FileName: entry.Name(),
			SQL:      string(content),
			Checksum: sha256.Sum256(content),
		})

		versions[version] = entry.Name()
		names[name] = entry.Name()
	}

	sort.Slice(result, func(first, second int) bool {
		return result[first].Version < result[second].Version
	})

	return result, nil
}

func parseFileName(fileName string) (uint32, string, error) {
	if path.Ext(fileName) != ".sql" {
		return 0, "", fmt.Errorf(
			"arquivo de migration %q deveria possuir extensão .sql",
			fileName,
		)
	}

	baseName := strings.TrimSuffix(fileName, ".sql")
	parts := strings.SplitN(baseName, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf(
			"arquivo de migration %q deveria seguir o formato NNNN_nome.sql",
			fileName,
		)
	}

	if len(parts[0]) != 4 {
		return 0, "", fmt.Errorf(
			"versão da migration %q deveria possuir quatro dígitos",
			fileName,
		)
	}

	version, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil || version == 0 {
		return 0, "", fmt.Errorf(
			"versão da migration %q é inválida",
			fileName,
		)
	}

	if !isValidName(parts[1]) {
		return 0, "", fmt.Errorf(
			"nome da migration %q é inválido",
			fileName,
		)
	}

	return uint32(version), parts[1], nil
}

func isValidName(value string) bool {
	if value == "" || value[0] < 'a' || value[0] > 'z' {
		return false
	}

	for _, character := range value {
		isLowercaseLetter := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'

		if !isLowercaseLetter && !isDigit && character != '_' {
			return false
		}
	}

	return true
}
