package publication

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestClassifyFileUsesEmbeddedCatalog(t *testing.T) {
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("não foi possível carregar o catálogo: %v", err)
	}

	tests := []struct {
		sourceName         string
		expectedDataset    string
		expectedPartNumber uint16
	}{
		{
			sourceName:         "Cnaes.zip",
			expectedDataset:    "economic_activities",
			expectedPartNumber: 0,
		},
		{
			sourceName:         "Empresas7.zip",
			expectedDataset:    "companies",
			expectedPartNumber: 7,
		},
		{
			sourceName:         "Estabelecimentos3.zip",
			expectedDataset:    "establishments",
			expectedPartNumber: 3,
		},
		{
			sourceName:         "Motivos.zip",
			expectedDataset:    "registration_status_reasons",
			expectedPartNumber: 0,
		},
		{
			sourceName:         "Municipios.zip",
			expectedDataset:    "municipalities",
			expectedPartNumber: 0,
		},
		{
			sourceName:         "Naturezas.zip",
			expectedDataset:    "legal_natures",
			expectedPartNumber: 0,
		},
		{
			sourceName:         "Paises.zip",
			expectedDataset:    "countries",
			expectedPartNumber: 0,
		},
		{
			sourceName:         "Qualificacoes.zip",
			expectedDataset:    "partner_qualifications",
			expectedPartNumber: 0,
		},
		{
			sourceName:         "Simples.zip",
			expectedDataset:    "simple_tax_options",
			expectedPartNumber: 0,
		},
		{
			sourceName:         "Socios9.zip",
			expectedDataset:    "partners",
			expectedPartNumber: 9,
		},
	}

	for _, test := range tests {
		t.Run(test.sourceName, func(t *testing.T) {
			sourceLocation := filepath.Join(
				"/dados/receita/2026-09",
				test.sourceName,
			)

			result, err := ClassifyFile(
				catalog,
				DiscoveredFile{
					SourceName:     test.sourceName,
					SourceLocation: sourceLocation,
				},
			)
			if err != nil {
				t.Fatalf("não esperava erro, mas recebeu: %v", err)
			}

			if result.DatasetCode != test.expectedDataset {
				t.Errorf(
					"dataset deveria ser %q, mas recebeu %q",
					test.expectedDataset,
					result.DatasetCode,
				)
			}

			if result.PartNumber != test.expectedPartNumber {
				t.Errorf(
					"parte deveria ser %d, mas recebeu %d",
					test.expectedPartNumber,
					result.PartNumber,
				)
			}

			if result.SourceName != test.sourceName {
				t.Errorf(
					"nome deveria ser %q, mas recebeu %q",
					test.sourceName,
					result.SourceName,
				)
			}

			if result.SourceLocation != sourceLocation {
				t.Errorf(
					"localização deveria ser %q, mas recebeu %q",
					sourceLocation,
					result.SourceLocation,
				)
			}
		})
	}
}

func TestClassifyFileRejectsInvalidFiles(t *testing.T) {
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("não foi possível carregar o catálogo: %v", err)
	}

	tests := []struct {
		name            string
		file            DiscoveredFile
		expectedMessage string
	}{
		{
			name: "nome vazio",
			file: DiscoveredFile{
				SourceLocation: "/dados/receita/arquivo.zip",
			},
			expectedMessage: "nome do arquivo descoberto é obrigatório",
		},
		{
			name: "localização vazia",
			file: DiscoveredFile{
				SourceName: "Cnaes.zip",
			},
			expectedMessage: "localização do arquivo descoberto é obrigatória",
		},
		{
			name: "arquivo desconhecido",
			file: DiscoveredFile{
				SourceName:     "ArquivoDesconhecido.zip",
				SourceLocation: "/dados/receita/ArquivoDesconhecido.zip",
			},
			expectedMessage: "não corresponde a nenhum dataset conhecido",
		},
		{
			name: "número da parte excede SMALLINT UNSIGNED",
			file: DiscoveredFile{
				SourceName:     "Empresas65536.zip",
				SourceLocation: "/dados/receita/Empresas65536.zip",
			},
			expectedMessage: "número de parte inválido",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ClassifyFile(catalog, test.file)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if result != (ClassifiedFile{}) {
				t.Errorf(
					"resultado deveria estar vazio, mas recebeu: %+v",
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

func TestClassifyFileRejectsInvalidCatalog(t *testing.T) {
	result, err := ClassifyFile(
		Catalog{},
		DiscoveredFile{
			SourceName:     "Cnaes.zip",
			SourceLocation: "/dados/receita/Cnaes.zip",
		},
	)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	if result != (ClassifiedFile{}) {
		t.Errorf(
			"resultado deveria estar vazio, mas recebeu: %+v",
			result,
		)
	}

	if !strings.Contains(err.Error(), "catálogo inválido") {
		t.Errorf(
			"erro deveria mencionar catálogo inválido, mas recebeu: %v",
			err,
		)
	}
}

func TestClassifyFileRejectsAmbiguousMatch(t *testing.T) {
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
				Code:        "special_companies",
				FilePattern: `^Empresas(7)\.zip$`,
				PartNumberRule: PartNumberRule{
					Source:       PartNumberSourceCaptureGroup,
					CaptureGroup: 1,
				},
			},
		},
	}

	result, err := ClassifyFile(
		catalog,
		DiscoveredFile{
			SourceName:     "Empresas7.zip",
			SourceLocation: "/dados/receita/Empresas7.zip",
		},
	)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	if result != (ClassifiedFile{}) {
		t.Errorf(
			"resultado deveria estar vazio, mas recebeu: %+v",
			result,
		)
	}

	if !strings.Contains(err.Error(), "mais de um dataset") {
		t.Errorf(
			"erro deveria mencionar correspondência ambígua, mas recebeu: %v",
			err,
		)
	}
}
