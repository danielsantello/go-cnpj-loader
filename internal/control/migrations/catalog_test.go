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

	expectedMigrations := []struct {
		version uint32
		name    string
	}{
		{
			version: 1,
			name:    "create_control_schema_migrations",
		},
		{
			version: 2,
			name:    "create_publications",
		},
		{
			version: 3,
			name:    "create_publication_files",
		},
		{
			version: 4,
			name:    "create_versions",
		},
		{
			version: 5,
			name:    "create_executions",
		},
		{
			version: 6,
			name:    "create_execution_configurations",
		},
		{
			version: 7,
			name:    "create_execution_steps",
		},
		{
			version: 8,
			name:    "create_file_loads",
		},
		{
			version: 9,
			name:    "create_execution_events",
		},
	}

	if len(result) != len(expectedMigrations) {
		t.Fatalf(
			"catálogo deveria possuir %d migrations, mas recebeu %d",
			len(expectedMigrations),
			len(result),
		)
	}

	for index, expected := range expectedMigrations {
		migration := result[index]

		if migration.Version != expected.version {
			t.Errorf(
				"migration %d deveria possuir versão %d, mas recebeu %d",
				index,
				expected.version,
				migration.Version,
			)
		}

		if migration.Name != expected.name {
			t.Errorf(
				"migration %d deveria possuir nome %q, mas recebeu %q",
				index,
				expected.name,
				migration.Name,
			)
		}

		if strings.TrimSpace(migration.SQL) == "" {
			t.Errorf(
				"conteúdo SQL da migration %d não deveria estar vazio",
				index,
			)
		}

		expectedChecksum := sha256.Sum256([]byte(migration.SQL))
		if migration.Checksum != expectedChecksum {
			t.Errorf(
				"checksum da migration %d deveria corresponder ao conteúdo SQL",
				index,
			)
		}
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
