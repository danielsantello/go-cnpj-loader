package publication

import (
	"strings"
	"testing"
)

func TestLoadCatalogReturnsEmbeddedDatasets(t *testing.T) {
	result, err := LoadCatalog()
	if err != nil {
		t.Fatalf("não foi possível carregar o catálogo: %v", err)
	}

	if result.FormatVersion != 1 {
		t.Errorf(
			"versão do formato deveria ser %d, mas recebeu %d",
			1,
			result.FormatVersion,
		)
	}

	expectedCodes := []string{
		"economic_activities",
		"companies",
		"establishments",
		"registration_status_reasons",
		"municipalities",
		"legal_natures",
		"countries",
		"partner_qualifications",
		"simple_tax_options",
		"partners",
	}

	if len(result.Datasets) != len(expectedCodes) {
		t.Fatalf(
			"catálogo deveria possuir %d datasets, mas recebeu %d",
			len(expectedCodes),
			len(result.Datasets),
		)
	}

	for index, expectedCode := range expectedCodes {
		dataset := result.Datasets[index]

		if dataset.Code != expectedCode {
			t.Errorf(
				"dataset %d deveria possuir código %q, mas recebeu %q",
				index,
				expectedCode,
				dataset.Code,
			)
		}
	}
}

func TestDecodeCatalogRejectsInvalidDocuments(t *testing.T) {
	validDocument := `{
		"format_version": 1,
		"datasets": [
			{
				"code": "countries",
				"file_pattern": "^Paises\\.zip$",
				"part_number": {
					"source": "fixed",
					"value": 0
				}
			}
		]
	}`

	tests := []struct {
		name            string
		document        string
		expectedMessage string
	}{
		{
			name:            "JSON malformado",
			document:        `{`,
			expectedMessage: "não foi possível decodificar",
		},
		{
			name: "campo desconhecido",
			document: `{
				"format_version": 1,
				"datasets": [],
				"unknown": true
			}`,
			expectedMessage: "não foi possível decodificar",
		},
		{
			name:            "dois documentos",
			document:        validDocument + `{}`,
			expectedMessage: "mais de um documento JSON",
		},
		{
			name:            "conteúdo adicional inválido",
			document:        validDocument + `invalid`,
			expectedMessage: "conteúdo adicional inválido",
		},
		{
			name: "semântica inválida",
			document: `{
				"format_version": 999,
				"datasets": []
			}`,
			expectedMessage: "catálogo de publicações inválido",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := decodeCatalog([]byte(test.document))
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if result.FormatVersion != 0 || result.Datasets != nil {
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
