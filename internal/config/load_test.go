package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadOverridesDefaultsWithEnvironmentVariables(t *testing.T) {
	t.Setenv(EnvEnvironment, "benchmark")
	t.Setenv(EnvMySQLHost, "mysql")
	t.Setenv(EnvMySQLPort, "3307")
	t.Setenv(EnvMySQLUser, "cnpj_loader")
	t.Setenv(EnvMySQLPassword, "secret")
	t.Setenv(EnvMySQLConnectTimeout, "15s")
	t.Setenv(EnvControlSchema, "custom_control")

	result, err := Load()
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	if result.Environment != EnvironmentBenchmark {
		t.Errorf(
			"ambiente deveria ser %q, mas recebeu %q",
			EnvironmentBenchmark,
			result.Environment,
		)
	}

	if result.MySQL.Host != "mysql" {
		t.Errorf(
			"host deveria ser %q, mas recebeu %q",
			"mysql",
			result.MySQL.Host,
		)
	}

	if result.MySQL.Port != 3307 {
		t.Errorf(
			"porta deveria ser %d, mas recebeu %d",
			3307,
			result.MySQL.Port,
		)
	}

	if result.MySQL.User != "cnpj_loader" {
		t.Errorf(
			"usuário deveria ser %q, mas recebeu %q",
			"cnpj_loader",
			result.MySQL.User,
		)
	}

	if result.MySQL.Password != "secret" {
		t.Errorf("senha não foi carregada corretamente")
	}

	if result.MySQL.ConnectTimeout != 15*time.Second {
		t.Errorf(
			"timeout deveria ser %s, mas recebeu %s",
			15*time.Second,
			result.MySQL.ConnectTimeout,
		)
	}

	if result.ControlSchema != "custom_control" {
		t.Errorf(
			"schema de controle deveria ser %q, mas recebeu %q",
			"custom_control",
			result.ControlSchema,
		)
	}
}

func TestLoadReturnsErrorForInvalidConvertedValues(t *testing.T) {
	tests := []struct {
		name             string
		environmentName  string
		environmentValue string
		expectedMessage  string
	}{
		{
			name:             "porta inválida",
			environmentName:  EnvMySQLPort,
			environmentValue: "invalid",
			expectedMessage:  EnvMySQLPort,
		},
		{
			name:             "timeout inválido",
			environmentName:  EnvMySQLConnectTimeout,
			environmentValue: "tomorrow",
			expectedMessage:  EnvMySQLConnectTimeout,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(test.environmentName, test.environmentValue)

			_, err := Load()
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
