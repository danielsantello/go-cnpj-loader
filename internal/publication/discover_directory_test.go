package publication

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverDirectoryReturnsZIPFiles(t *testing.T) {
	directory := t.TempDir()

	fileNames := []string{
		"Empresas1.zip",
		"Cnaes.ZIP",
		"README.txt",
	}

	for _, fileName := range fileNames {
		filePath := filepath.Join(directory, fileName)

		if err := os.WriteFile(
			filePath,
			[]byte("test"),
			0o600,
		); err != nil {
			t.Fatalf(
				"não foi possível criar o arquivo de teste %q: %v",
				fileName,
				err,
			)
		}
	}

	if err := os.Mkdir(
		filepath.Join(directory, "Ignorado.zip"),
		0o700,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o subdiretório de teste: %v",
			err,
		)
	}

	result, err := DiscoverDirectory(
		Source{
			Type:     SourceTypeDirectory,
			Location: directory,
		},
	)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	expected := []DiscoveredFile{
		{
			SourceName:     "Cnaes.ZIP",
			SourceLocation: filepath.Join(directory, "Cnaes.ZIP"),
		},
		{
			SourceName:     "Empresas1.zip",
			SourceLocation: filepath.Join(directory, "Empresas1.zip"),
		},
	}

	if len(result) != len(expected) {
		t.Fatalf(
			"deveria encontrar %d arquivos ZIP, mas encontrou %d",
			len(expected),
			len(result),
		)
	}

	for index, expectedFile := range expected {
		if result[index] != expectedFile {
			t.Errorf(
				"arquivo %d deveria ser %+v, mas recebeu %+v",
				index,
				expectedFile,
				result[index],
			)
		}
	}
}

func TestDiscoverDirectoryRejectsInvalidSources(t *testing.T) {
	emptyDirectory := t.TempDir()
	missingDirectory := filepath.Join(t.TempDir(), "missing")

	tests := []struct {
		name            string
		source          Source
		expectedMessage string
	}{
		{
			name: "tipo URL",
			source: Source{
				Type:     SourceTypeURL,
				Location: "https://example.com/cnpj",
			},
			expectedMessage: "origem deveria ser do tipo",
		},
		{
			name: "localização vazia",
			source: Source{
				Type: SourceTypeDirectory,
			},
			expectedMessage: "localização do diretório da publicação é obrigatória",
		},
		{
			name: "caminho relativo",
			source: Source{
				Type:     SourceTypeDirectory,
				Location: "dados/receita",
			},
			expectedMessage: "deve ser absoluta",
		},
		{
			name: "diretório inexistente",
			source: Source{
				Type:     SourceTypeDirectory,
				Location: missingDirectory,
			},
			expectedMessage: "não foi possível ler o diretório",
		},
		{
			name: "diretório sem arquivos ZIP",
			source: Source{
				Type:     SourceTypeDirectory,
				Location: emptyDirectory,
			},
			expectedMessage: "não possui arquivos ZIP",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := DiscoverDirectory(test.source)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if result != nil {
				t.Errorf(
					"resultado deveria ser nil, mas recebeu: %+v",
					result,
				)
			}

			if !strings.Contains(err.Error(), test.expectedMessage) {
				t.Errorf(
					"erro deveria mencionar %q, mas recebeu: %v",
					test.expectedMessage,
					err,
				)
			}
		})
	}
}
