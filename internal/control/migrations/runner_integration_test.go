package migrations

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-loader/internal/buildinfo"
	"github.com/danielsantello/go-cnpj-loader/internal/config"
	"github.com/danielsantello/go-cnpj-loader/internal/database"
)

const (
	runIntegrationTestsEnvironment = "CNPJ_LOADER_RUN_INTEGRATION_TESTS"
	integrationTestSchema          = "cnpj_loader_migrations_integration_test"
)

func TestBootstrapCreatesControlSchema(t *testing.T) {
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

	connection, err := database.OpenMySQL(value.MySQL)
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
		30*time.Second,
	)
	defer cancel()

	if err := database.PingMySQL(ctx, connection); err != nil {
		t.Fatalf("falha ao validar a conexão: %v", err)
	}

	var schemaExists bool
	err = connection.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.SCHEMATA
				WHERE SCHEMA_NAME = ?
			)
		`,
		integrationTestSchema,
	).Scan(&schemaExists)
	if err != nil {
		t.Fatalf(
			"não foi possível verificar o schema temporário: %v",
			err,
		)
	}

	if schemaExists {
		t.Fatalf(
			"schema temporário %q já existe; remova-o após investigar",
			integrationTestSchema,
		)
	}

	t.Cleanup(func() {
		statement := fmt.Sprintf(
			"DROP DATABASE IF EXISTS `%s`",
			integrationTestSchema,
		)

		if _, err := connection.ExecContext(
			context.Background(),
			statement,
		); err != nil {
			t.Errorf(
				"não foi possível remover o schema temporário %q: %v",
				integrationTestSchema,
				err,
			)
		}
	})

	build := buildinfo.Info{
		Version:   "integration-test",
		Commit:    "integration-test",
		BuildDate: "unknown",
		GoVersion: "integration-test",
	}

	if err := Bootstrap(
		ctx,
		connection,
		integrationTestSchema,
		build,
	); err != nil {
		t.Fatalf(
			"não foi possível executar o bootstrap: %v",
			err,
		)
	}

	var tableExists bool
	err = connection.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.TABLES
				WHERE TABLE_SCHEMA = ?
					AND TABLE_NAME = 'control_schema_migrations'
			)
		`,
		integrationTestSchema,
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf(
			"não foi possível verificar a tabela de migrations: %v",
			err,
		)
	}

	if !tableExists {
		t.Fatal("tabela de migrations deveria existir")
	}

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("não foi possível carregar o catálogo: %v", err)
	}

	var (
		version       uint32
		name          string
		checksum      []byte
		status        string
		loaderVersion string
		loaderCommit  string
	)

	query := fmt.Sprintf(
		`
			SELECT
				version,
				name,
				checksum,
				status,
				loader_version,
				loader_commit
			FROM %s.control_schema_migrations
			WHERE version = 1
		`,
		integrationTestSchema,
	)

	err = connection.QueryRowContext(ctx, query).Scan(
		&version,
		&name,
		&checksum,
		&status,
		&loaderVersion,
		&loaderCommit,
	)
	if err != nil {
		t.Fatalf(
			"não foi possível consultar a migration registrada: %v",
			err,
		)
	}

	expectedMigration := catalog[0]

	if version != expectedMigration.Version {
		t.Errorf(
			"versão deveria ser %d, mas recebeu %d",
			expectedMigration.Version,
			version,
		)
	}

	if name != expectedMigration.Name {
		t.Errorf(
			"nome deveria ser %q, mas recebeu %q",
			expectedMigration.Name,
			name,
		)
	}

	if !bytes.Equal(checksum, expectedMigration.Checksum[:]) {
		t.Error("checksum registrado deveria corresponder ao catálogo")
	}

	if status != "applied" {
		t.Errorf(
			"status deveria ser %q, mas recebeu %q",
			"applied",
			status,
		)
	}

	if loaderVersion != build.Version {
		t.Errorf(
			"versão do loader deveria ser %q, mas recebeu %q",
			build.Version,
			loaderVersion,
		)
	}

	if loaderCommit != build.Commit {
		t.Errorf(
			"commit do loader deveria ser %q, mas recebeu %q",
			build.Commit,
			loaderCommit,
		)
	}
}
