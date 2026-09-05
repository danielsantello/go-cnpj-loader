package publication

import (
	"strings"
	"testing"
)

func TestValidateCatalogAcceptsValidCatalog(t *testing.T) {
	value := Catalog{
		FormatVersion: CurrentCatalogFormatVersion,
		Datasets: []Dataset{
			{
				Code:        "countries",
				FilePattern: `^Paises\.zip$`,
				PartNumberRule: PartNumberRule{
					Source: PartNumberSourceFixed,
					Value:  0,
				},
			},
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

	err := ValidateCatalog(value)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
}

func TestValidateCatalogRejectsInvalidCatalogs(t *testing.T) {
	newValidCatalog := func() Catalog {
		return Catalog{
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
	}

	tests := []struct {
		name            string
		change          func(*Catalog)
		expectedMessage string
	}{
		{
			name: "versão desconhecida",
			change: func(value *Catalog) {
				value.FormatVersion = 999
			},
			expectedMessage: "versão do formato",
		},
		{
			name: "sem datasets",
			change: func(value *Catalog) {
				value.Datasets = nil
			},
			expectedMessage: "não possui datasets",
		},
		{
			name: "código inválido",
			change: func(value *Catalog) {
				value.Datasets[0].Code = "Invalid-Code"
			},
			expectedMessage: "código inválido",
		},
		{
			name: "código duplicado",
			change: func(value *Catalog) {
				secondDataset := value.Datasets[0]
				secondDataset.FilePattern = `^OutrasEmpresas([0-9]+)\.zip$`
				value.Datasets = append(value.Datasets, secondDataset)
			},
			expectedMessage: "código duplicado",
		},
		{
			name: "padrão vazio",
			change: func(value *Catalog) {
				value.Datasets[0].FilePattern = ""
			},
			expectedMessage: "não possui padrão de arquivo",
		},
		{
			name: "padrão duplicado",
			change: func(value *Catalog) {
				secondDataset := value.Datasets[0]
				secondDataset.Code = "other_companies"
				value.Datasets = append(value.Datasets, secondDataset)
			},
			expectedMessage: "padrão de arquivo duplicado",
		},
		{
			name: "padrão sem âncoras",
			change: func(value *Catalog) {
				value.Datasets[0].FilePattern = `Empresas([0-9]+)\.zip`
			},
			expectedMessage: "padrão de arquivo ancorado",
		},
		{
			name: "expressão regular inválida",
			change: func(value *Catalog) {
				value.Datasets[0].FilePattern = `^([$`
			},
			expectedMessage: "padrão de arquivo inválido",
		},
		{
			name: "grupo de captura inexistente",
			change: func(value *Catalog) {
				value.Datasets[0].PartNumberRule.CaptureGroup = 2
			},
			expectedMessage: "grupo de captura inválido",
		},
		{
			name: "valor fixo com parte capturada",
			change: func(value *Catalog) {
				value.Datasets[0].PartNumberRule.Value = 1
			},
			expectedMessage: "não pode definir valor fixo",
		},
		{
			name: "parte fixa com grupo de captura",
			change: func(value *Catalog) {
				value.Datasets[0].PartNumberRule.Source = PartNumberSourceFixed
			},
			expectedMessage: "parte fixa não pode definir grupo de captura",
		},
		{
			name: "origem da parte desconhecida",
			change: func(value *Catalog) {
				value.Datasets[0].PartNumberRule.Source = "unknown"
			},
			expectedMessage: "origem do número da parte inválida",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := newValidCatalog()
			test.change(&value)

			err := ValidateCatalog(value)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
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
