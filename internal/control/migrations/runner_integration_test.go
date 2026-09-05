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

	if err := Migrate(
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

	if err := Migrate(
		ctx,
		connection,
		integrationTestSchema,
		build,
	); err != nil {
		t.Fatalf(
			"não foi possível executar novamente as migrations: %v",
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

	err = connection.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.TABLES
				WHERE TABLE_SCHEMA = ?
					AND TABLE_NAME = 'publications'
			)
		`,
		integrationTestSchema,
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf(
			"não foi possível verificar a tabela de publicações: %v",
			err,
		)
	}

	if !tableExists {
		t.Fatal("tabela de publicações deveria existir")
	}

	err = connection.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.TABLES
				WHERE TABLE_SCHEMA = ?
					AND TABLE_NAME = 'publication_files'
			)
		`,
		integrationTestSchema,
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf(
			"não foi possível verificar a tabela de arquivos da publicação: %v",
			err,
		)
	}

	if !tableExists {
		t.Fatal("tabela de arquivos da publicação deveria existir")
	}

	err = connection.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.TABLES
				WHERE TABLE_SCHEMA = ?
					AND TABLE_NAME = 'versions'
			)
		`,
		integrationTestSchema,
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf(
			"não foi possível verificar a tabela de versões: %v",
			err,
		)
	}

	if !tableExists {
		t.Fatal("tabela de versões deveria existir")
	}

	err = connection.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.TABLES
				WHERE TABLE_SCHEMA = ?
					AND TABLE_NAME = 'executions'
			)
		`,
		integrationTestSchema,
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf(
			"não foi possível verificar a tabela de execuções: %v",
			err,
		)
	}

	if !tableExists {
		t.Fatal("tabela de execuções deveria existir")
	}

	err = connection.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.TABLES
				WHERE TABLE_SCHEMA = ?
					AND TABLE_NAME = 'execution_configurations'
			)
		`,
		integrationTestSchema,
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf(
			"não foi possível verificar a tabela de configurações das execuções: %v",
			err,
		)
	}

	if !tableExists {
		t.Fatal("tabela de configurações das execuções deveria existir")
	}

	err = connection.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.TABLES
				WHERE TABLE_SCHEMA = ?
					AND TABLE_NAME = 'execution_steps'
			)
		`,
		integrationTestSchema,
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf(
			"não foi possível verificar a tabela de etapas das execuções: %v",
			err,
		)
	}

	if !tableExists {
		t.Fatal("tabela de etapas das execuções deveria existir")
	}

	err = connection.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.TABLES
				WHERE TABLE_SCHEMA = ?
					AND TABLE_NAME = 'file_loads'
			)
		`,
		integrationTestSchema,
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf(
			"não foi possível verificar a tabela de cargas dos arquivos: %v",
			err,
		)
	}

	if !tableExists {
		t.Fatal("tabela de cargas dos arquivos deveria existir")
	}

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("não foi possível carregar o catálogo: %v", err)
	}

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
			ORDER BY version
		`,
		integrationTestSchema,
	)

	rows, err := connection.QueryContext(ctx, query)
	if err != nil {
		t.Fatalf(
			"não foi possível consultar as migrations registradas: %v",
			err,
		)
	}
	defer rows.Close()

	migrationIndex := 0

	for rows.Next() {
		if migrationIndex >= len(catalog) {
			t.Fatal("histórico possui mais migrations que o catálogo")
		}

		var (
			version       uint32
			name          string
			checksum      []byte
			status        string
			loaderVersion string
			loaderCommit  string
		)

		if err := rows.Scan(
			&version,
			&name,
			&checksum,
			&status,
			&loaderVersion,
			&loaderCommit,
		); err != nil {
			t.Fatalf(
				"não foi possível ler a migration registrada: %v",
				err,
			)
		}

		expectedMigration := catalog[migrationIndex]

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
			t.Errorf(
				"checksum da migration %d deveria corresponder ao catálogo",
				expectedMigration.Version,
			)
		}

		if status != "applied" {
			t.Errorf(
				"status da migration %d deveria ser %q, mas recebeu %q",
				expectedMigration.Version,
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

		migrationIndex++
	}

	if err := rows.Err(); err != nil {
		t.Fatalf(
			"não foi possível percorrer as migrations registradas: %v",
			err,
		)
	}

	if migrationIndex != len(catalog) {
		t.Fatalf(
			"histórico deveria possuir %d migrations, mas recebeu %d",
			len(catalog),
			migrationIndex,
		)
	}

	migrationSession, err := connection.Conn(ctx)
	if err != nil {
		t.Fatalf(
			"não foi possível obter sessão para testar falha: %v",
			err,
		)
	}
	defer migrationSession.Close()

	useSchemaStatement := fmt.Sprintf(
		"USE `%s`",
		integrationTestSchema,
	)

	if _, err := migrationSession.ExecContext(
		ctx,
		useSchemaStatement,
	); err != nil {
		t.Fatalf(
			"não foi possível selecionar o schema temporário: %v",
			err,
		)
	}

	failingMigration := Migration{
		Version:  4294967295,
		Name:     "integration_test_failure",
		FileName: "integration_test_failure.sql",
		SQL:      "CREATE TABLE publications (id INT)",
	}

	err = applyTrackedMigration(
		ctx,
		migrationSession,
		failingMigration,
		build,
	)
	if err == nil {
		t.Fatal("migration inválida deveria retornar erro")
	}

	var (
		failedStatus   string
		mysqlErrorCode uint16
		sqlState       string
		errorMessage   string
		finishedAt     time.Time
	)

	failureQuery := fmt.Sprintf(
		`
			SELECT
				status,
				mysql_error_code,
				sql_state,
				error_message,
				finished_at_utc
			FROM %s.control_schema_migrations
			WHERE version = ?
		`,
		integrationTestSchema,
	)

	err = connection.QueryRowContext(
		ctx,
		failureQuery,
		failingMigration.Version,
	).Scan(
		&failedStatus,
		&mysqlErrorCode,
		&sqlState,
		&errorMessage,
		&finishedAt,
	)
	if err != nil {
		t.Fatalf(
			"não foi possível consultar a migration com falha: %v",
			err,
		)
	}

	if failedStatus != "failed" {
		t.Errorf(
			"status deveria ser %q, mas recebeu %q",
			"failed",
			failedStatus,
		)
	}

	if mysqlErrorCode != 1050 {
		t.Errorf(
			"código MySQL deveria ser %d, mas recebeu %d",
			1050,
			mysqlErrorCode,
		)
	}

	if sqlState != "42S01" {
		t.Errorf(
			"SQL state deveria ser %q, mas recebeu %q",
			"42S01",
			sqlState,
		)
	}

	if errorMessage == "" {
		t.Error("mensagem de erro não deveria estar vazia")
	}

	if finishedAt.IsZero() {
		t.Error("data de conclusão da falha não deveria estar vazia")
	}
}
