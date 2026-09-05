package publication

import "testing"

func TestParseSourceReturnsURL(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "HTTPS",
			value: "https://example.com/cnpj/2026-09",
		},
		{
			name:  "HTTP",
			value: "http://example.com/cnpj/2026-09",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseSource(test.value)
			if err != nil {
				t.Fatalf("não esperava erro, mas recebeu: %v", err)
			}

			if result.Type != SourceTypeURL {
				t.Errorf(
					"tipo deveria ser %q, mas recebeu %q",
					SourceTypeURL,
					result.Type,
				)
			}

			if result.Location != test.value {
				t.Errorf(
					"localização deveria ser %q, mas recebeu %q",
					test.value,
					result.Location,
				)
			}
		})
	}
}

func TestParseSourceReturnsDirectory(t *testing.T) {
	value := "/dados/receita/2026-09"

	result, err := ParseSource(value)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	if result.Type != SourceTypeDirectory {
		t.Errorf(
			"tipo deveria ser %q, mas recebeu %q",
			SourceTypeDirectory,
			result.Type,
		)
	}

	if result.Location != value {
		t.Errorf(
			"localização deveria ser %q, mas recebeu %q",
			value,
			result.Location,
		)
	}
}

func TestParseSourceRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "valor vazio",
			value: "",
		},
		{
			name:  "somente espaços",
			value: "   ",
		},
		{
			name:  "caminho relativo",
			value: "dados/receita/2026-09",
		},
		{
			name:  "espaço externo",
			value: " /dados/receita/2026-09",
		},
		{
			name:  "protocolo não permitido",
			value: "ftp://example.com/cnpj",
		},
		{
			name:  "URL sem host",
			value: "https://",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseSource(test.value)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if result != (Source{}) {
				t.Errorf(
					"resultado deveria estar vazio, mas recebeu: %+v",
					result,
				)
			}
		})
	}
}
