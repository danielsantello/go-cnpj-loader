package database

import (
	"context"
	"os"
	"testing"

	"github.com/danielsantello/go-cnpj-loader/internal/config"
)

const runIntegrationTestsEnvironment = "CNPJ_LOADER_RUN_INTEGRATION_TESTS"

func TestMySQLConnection(t *testing.T) {
	if os.Getenv(runIntegrationTestsEnvironment) != "1" {
		t.Skip("teste de integração com MySQL não solicitado")
	}

	value, err := config.Load()
	if err != nil {
		t.Fatalf("não foi possível carregar a configuração: %v", err)
	}

	if err := config.Validate(value); err != nil {
		t.Fatalf("configuração inválida: %v", err)
	}

	connection, err := OpenMySQL(value.MySQL)
	if err != nil {
		t.Fatalf("não foi possível abrir o gerenciador de conexões: %v", err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("não foi possível fechar a conexão: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		value.MySQL.ConnectTimeout,
	)
	defer cancel()

	if err := PingMySQL(ctx, connection); err != nil {
		t.Fatalf("falha ao validar a conexão: %v", err)
	}
}
