package config

import (
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	result := Default()

	if result.Environment != EnvironmentDevelopment {
		t.Errorf(
			"ambiente padrão deveria ser %q, mas recebeu %q",
			EnvironmentDevelopment,
			result.Environment,
		)
	}

	if result.MySQL.Host != "127.0.0.1" {
		t.Errorf(
			"host padrão deveria ser %q, mas recebeu %q",
			"127.0.0.1",
			result.MySQL.Host,
		)
	}

	if result.MySQL.Port != 3306 {
		t.Errorf(
			"porta padrão deveria ser %d, mas recebeu %d",
			3306,
			result.MySQL.Port,
		)
	}

	if result.MySQL.ConnectTimeout != 5*time.Second {
		t.Errorf(
			"timeout padrão deveria ser %s, mas recebeu %s",
			5*time.Second,
			result.MySQL.ConnectTimeout,
		)
	}

	if result.ControlSchema != "cnpj_loader_control" {
		t.Errorf(
			"schema de controle padrão deveria ser %q, mas recebeu %q",
			"cnpj_loader_control",
			result.ControlSchema,
		)
	}

	if result.WorkspacePath != "" {
		t.Errorf(
			"workspace não deveria possuir valor padrão, mas recebeu %q",
			result.WorkspacePath,
		)
	}

	if result.MySQL.User != "" {
		t.Errorf("usuário do MySQL não deveria possuir valor padrão")
	}

	if result.MySQL.Password != "" {
		t.Errorf("senha do MySQL não deveria possuir valor padrão")
	}
}
