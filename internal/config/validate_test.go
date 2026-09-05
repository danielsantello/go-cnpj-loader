package config

import (
	"strings"
	"testing"
)

func TestValidateAcceptsValidConfiguration(t *testing.T) {
	value := Default()
	value.MySQL.User = "cnpj_loader"
	value.MySQL.Password = "secret"

	err := Validate(value)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
}

func TestValidateReturnsAllConfigurationProblems(t *testing.T) {
	value := Config{
		Environment: "invalid",
		MySQL: MySQL{
			Host:           " ",
			Port:           70000,
			ConnectTimeout: 0,
		},
		ControlSchema: "Invalid-Schema",
	}

	err := Validate(value)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	expectedMessages := []string{
		EnvEnvironment,
		EnvMySQLHost,
		EnvMySQLPort,
		EnvMySQLUser,
		EnvMySQLPassword,
		EnvMySQLConnectTimeout,
		EnvControlSchema,
	}

	for _, expectedMessage := range expectedMessages {
		if !strings.Contains(err.Error(), expectedMessage) {
			t.Errorf(
				"erro deveria mencionar %q, mas recebeu:\n%v",
				expectedMessage,
				err,
			)
		}
	}
}

func TestValidateControlSchemaName(t *testing.T) {
	tests := []struct {
		name       string
		schemaName string
		valid      bool
	}{
		{
			name:       "nome padrão",
			schemaName: "cnpj_loader_control",
			valid:      true,
		},
		{
			name:       "nome mínimo",
			schemaName: "a",
			valid:      true,
		},
		{
			name:       "nome com sessenta e quatro caracteres",
			schemaName: "a" + strings.Repeat("b", 63),
			valid:      true,
		},
		{
			name:       "nome vazio",
			schemaName: "",
			valid:      false,
		},
		{
			name:       "inicia com número",
			schemaName: "1cnpj",
			valid:      false,
		},
		{
			name:       "contém letra maiúscula",
			schemaName: "CnpjLoader",
			valid:      false,
		},
		{
			name:       "contém hífen",
			schemaName: "cnpj-loader",
			valid:      false,
		},
		{
			name:       "possui sessenta e cinco caracteres",
			schemaName: "a" + strings.Repeat("b", 64),
			valid:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := Default()
			value.MySQL.User = "cnpj_loader"
			value.MySQL.Password = "secret"
			value.ControlSchema = test.schemaName

			err := Validate(value)

			if test.valid && err != nil {
				t.Fatalf("não esperava erro, mas recebeu: %v", err)
			}

			if !test.valid && err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if !test.valid && !strings.Contains(err.Error(), EnvControlSchema) {
				t.Errorf(
					"erro deveria mencionar %q, mas recebeu: %v",
					EnvControlSchema,
					err,
				)
			}
		})
	}
}

func TestValidateLoadAcceptsAbsoluteWorkspacePath(t *testing.T) {
	value := Default()
	value.MySQL.User = "cnpj_loader"
	value.MySQL.Password = "secret"
	value.WorkspacePath = "/dados/cnpj-loader"

	err := ValidateLoad(value)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
}

func TestValidateLoadRejectsInvalidWorkspacePath(t *testing.T) {
	tests := []struct {
		name          string
		workspacePath string
	}{
		{
			name:          "caminho vazio",
			workspacePath: "",
		},
		{
			name:          "caminho relativo",
			workspacePath: "workspace",
		},
		{
			name:          "caminho absoluto com espaço inicial",
			workspacePath: " /dados/cnpj-loader",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := Default()
			value.MySQL.User = "cnpj_loader"
			value.MySQL.Password = "secret"
			value.WorkspacePath = test.workspacePath

			err := ValidateLoad(value)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if !strings.Contains(err.Error(), EnvWorkspace) {
				t.Errorf(
					"erro deveria mencionar %q, mas recebeu: %v",
					EnvWorkspace,
					err,
				)
			}
		})
	}
}
