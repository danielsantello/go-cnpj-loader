package publication

import (
	"strings"
	"testing"
)

func TestValidatePublicationFilesAcceptsCompletePublication(t *testing.T) {
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

	files := []ClassifiedFile{
		{
			DatasetCode:    "companies",
			PartNumber:     0,
			SourceName:     "Empresas0.zip",
			SourceLocation: "/dados/receita/Empresas0.zip",
		},
		{
			DatasetCode:    "companies",
			PartNumber:     1,
			SourceName:     "Empresas1.zip",
			SourceLocation: "/dados/receita/Empresas1.zip",
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

	err := ValidatePublicationFiles(catalog, files)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
}

func TestValidatePublicationFilesRejectsInvalidInputs(t *testing.T) {
	validCatalog := Catalog{
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
		},
	}

	tests := []struct {
		name            string
		catalog         Catalog
		files           []ClassifiedFile
		expectedMessage string
	}{
		{
			name:            "catálogo inválido",
			catalog:         Catalog{},
			files:           []ClassifiedFile{{}},
			expectedMessage: "catálogo inválido",
		},
		{
			name:            "lista vazia",
			catalog:         validCatalog,
			files:           nil,
			expectedMessage: "não possui arquivos classificados",
		},
		{
			name:    "dataset desconhecido",
			catalog: validCatalog,
			files: []ClassifiedFile{
				{
					DatasetCode: "unknown",
					PartNumber:  0,
					SourceName:  "Desconhecido.zip",
				},
			},
			expectedMessage: "dataset desconhecido",
		},
		{
			name:    "parte duplicada",
			catalog: validCatalog,
			files: []ClassifiedFile{
				{
					DatasetCode: "companies",
					PartNumber:  0,
					SourceName:  "Empresas0.zip",
				},
				{
					DatasetCode: "companies",
					PartNumber:  0,
					SourceName:  "OutraEmpresas0.zip",
				},
			},
			expectedMessage: "parte 0 duplicada",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePublicationFiles(
				test.catalog,
				test.files,
			)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
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

func TestValidatePublicationFilesRejectsIncompletePublication(t *testing.T) {
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

	tests := []struct {
		name            string
		files           []ClassifiedFile
		expectedMessage string
	}{
		{
			name: "dataset ausente",
			files: []ClassifiedFile{
				{
					DatasetCode: "companies",
					PartNumber:  0,
					SourceName:  "Empresas0.zip",
				},
			},
			expectedMessage: `não possui arquivos do dataset "countries"`,
		},
		{
			name: "dataset de parte fixa com várias partes",
			files: []ClassifiedFile{
				{
					DatasetCode: "companies",
					PartNumber:  0,
					SourceName:  "Empresas0.zip",
				},
				{
					DatasetCode: "countries",
					PartNumber:  0,
					SourceName:  "Paises.zip",
				},
				{
					DatasetCode: "countries",
					PartNumber:  1,
					SourceName:  "OutroPaises.zip",
				},
			},
			expectedMessage: "de parte fixa deveria possuir somente a parte 0",
		},
		{
			name: "parte fixa incorreta",
			files: []ClassifiedFile{
				{
					DatasetCode: "companies",
					PartNumber:  0,
					SourceName:  "Empresas0.zip",
				},
				{
					DatasetCode: "countries",
					PartNumber:  1,
					SourceName:  "Paises.zip",
				},
			},
			expectedMessage: "não possui a parte fixa esperada 0",
		},
		{
			name: "lacuna nas partes capturadas",
			files: []ClassifiedFile{
				{
					DatasetCode: "companies",
					PartNumber:  0,
					SourceName:  "Empresas0.zip",
				},
				{
					DatasetCode: "companies",
					PartNumber:  2,
					SourceName:  "Empresas2.zip",
				},
				{
					DatasetCode: "countries",
					PartNumber:  0,
					SourceName:  "Paises.zip",
				},
			},
			expectedMessage: `dataset "companies" não possui a parte esperada 1`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePublicationFiles(catalog, test.files)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
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
