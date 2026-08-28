package migrations

import (
	"crypto/sha256"
	"strings"
	"testing"
)

func TestLoadCatalogReturnsEmbeddedMigrations(t *testing.T) {
	result, err := LoadCatalog()
	if err != nil {
		t.Fatalf("não foi possível carregar o catálogo: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf(
			"catálogo deveria possuir %d migration, mas recebeu %d",
			1,
			len(result),
		)
	}

	migration := result[0]

	if migration.Version != 1 {
		t.Errorf(
			"versão deveria ser %d, mas recebeu %d",
			1,
			migration.Version,
		)
	}

	if migration.Name != "create_control_schema_migrations" {
		t.Errorf(
			"nome deveria ser %q, mas recebeu %q",
			"create_control_schema_migrations",
			migration.Name,
		)
	}

	if strings.TrimSpace(migration.SQL) == "" {
		t.Error("conteúdo SQL da migration não deveria estar vazio")
	}

	expectedChecksum := sha256.Sum256([]byte(migration.SQL))
	if migration.Checksum != expectedChecksum {
		t.Error("checksum deveria corresponder ao conteúdo SQL")
	}
}

func TestParseFileName(t *testing.T) {
	version, name, err := parseFileName(
		"0010_create_data_loads.sql",
	)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	if version != 10 {
		t.Errorf(
			"versão deveria ser %d, mas recebeu %d",
			10,
			version,
		)
	}

	if name != "create_data_loads" {
		t.Errorf(
			"nome deveria ser %q, mas recebeu %q",
			"create_data_loads",
			name,
		)
	}
}

func TestParseFileNameRejectsInvalidNames(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
	}{
		{
			name:     "sem extensão SQL",
			fileName: "0001_create_table.txt",
		},
		{
			name:     "sem separador",
			fileName: "0001.sql",
		},
		{
			name:     "versão com menos de quatro dígitos",
			fileName: "001_create_table.sql",
		},
		{
			name:     "versão igual a zero",
			fileName: "0000_create_table.sql",
		},
		{
			name:     "versão não numérica",
			fileName: "abcd_create_table.sql",
		},
		{
			name:     "nome vazio",
			fileName: "0001_.sql",
		},
		{
			name:     "nome começando com número",
			fileName: "0001_1create_table.sql",
		},
		{
			name:     "nome com letra maiúscula",
			fileName: "0001_Create_table.sql",
		},
		{
			name:     "nome com hífen",
			fileName: "0001_create-table.sql",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := parseFileName(test.fileName)
			if err == nil {
				t.Errorf(
					"esperava erro para o arquivo %q, mas recebeu nil",
					test.fileName,
				)
			}
		})
	}
}
