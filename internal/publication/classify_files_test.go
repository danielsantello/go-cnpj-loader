package publication

import (
	"strings"
	"testing"
)

func TestClassifyFilesReturnsCatalogOrder(t *testing.T) {
	catalog := Catalog{
		FormatVersion: CurrentCatalogFormatVersion,
		Datasets: []Dataset{
			{
				Code:        "companies",
				FilePattern: `^Empresas([0-9]+)\.zip$`,
				PartNumberRule: PartNumberRule{
					Source:       PartNumberSourceCaptureGroup,
					CaptureGroup: 1,
				},
			},
			{
				Code:        "countries",
				FilePattern: `^Paises\.zip$`,
				PartNumberRule: PartNumberRule{
					Source: PartNumberSourceFixed,
					Value:  0,
				},
			},
		},
	}

	files := []DiscoveredFile{
		{
			SourceName:     "Paises.zip",
			SourceLocation: "/dados/receita/Paises.zip",
		},
		{
			SourceName:     "Empresas2.zip",
			SourceLocation: "/dados/receita/Empresas2.zip",
		},
		{
			SourceName:     "Empresas0.zip",
			SourceLocation: "/dados/receita/Empresas0.zip",
		},
	}

	result, err := ClassifyFiles(catalog, files)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	expected := []ClassifiedFile{
		{
			DatasetCode:    "companies",
			PartNumber:     0,
			SourceName:     "Empresas0.zip",
			SourceLocation: "/dados/receita/Empresas0.zip",
		},
		{
			DatasetCode:    "companies",
			PartNumber:     2,
			SourceName:     "Empresas2.zip",
			SourceLocation: "/dados/receita/Empresas2.zip",
		},
		{
			DatasetCode:    "countries",
			PartNumber:     0,
			SourceName:     "Paises.zip",
			SourceLocation: "/dados/receita/Paises.zip",
		},
	}

	if len(result) != len(expected) {
		t.Fatalf(
			"resultado deveria possuir %d arquivos, mas recebeu %d",
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

func TestClassifyFilesRejectsInvalidCollections(t *testing.T) {
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("não foi possível carregar o catálogo: %v", err)
	}

	tests := []struct {
		name            string
		files           []DiscoveredFile
		expectedMessage string
	}{
		{
			name:            "lista vazia",
			files:           nil,
			expectedMessage: "não existem arquivos",
		},
		{
			name: "arquivo desconhecido",
			files: []DiscoveredFile{
				{
					SourceName:     "Desconhecido.zip",
					SourceLocation: "/dados/receita/Desconhecido.zip",
				},
			},
			expectedMessage: "não foi possível classificar",
		},
		{
			name: "dataset e parte duplicados",
			files: []DiscoveredFile{
				{
					SourceName:     "Empresas0.zip",
					SourceLocation: "/dados/origem-a/Empresas0.zip",
				},
				{
					SourceName:     "Empresas0.zip",
					SourceLocation: "/dados/origem-b/Empresas0.zip",
				},
			},
			expectedMessage: "representam o mesmo dataset",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ClassifyFiles(catalog, test.files)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if result != nil {
				t.Errorf(
					"resultado deveria ser nil, mas recebeu: %#v",
					result,
				)
			}

			if !strings.Contains(err.Error(), test.expectedMessage) {
				t.Errorf(
					"erro deveria conter %q, mas recebeu: %v",
					test.expectedMessage,
					err,
				)
			}
		})
	}
}
