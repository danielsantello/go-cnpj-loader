package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/danielsantello/go-cnpj-loader/internal/buildinfo"
)

var schemaNamePattern = regexp.MustCompile(
	`^[a-z][a-z0-9_]{0,63}$`,
)

func Bootstrap(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	build buildinfo.Info,
) error {
	if connection == nil {
		return fmt.Errorf("conexão com o MySQL não foi informada")
	}

	if !schemaNamePattern.MatchString(schemaName) {
		return fmt.Errorf(
			"nome do schema de controle é inválido: %q",
			schemaName,
		)
	}

	catalog, err := LoadCatalog()
	if err != nil {
		return fmt.Errorf(
			"não foi possível carregar o catálogo de migrations: %w",
			err,
		)
	}

	initialMigration := catalog[0]
	if initialMigration.Version != 1 {
		return fmt.Errorf(
			"primeira migration deveria possuir versão 1, mas recebeu %d",
			initialMigration.Version,
		)
	}

	session, err := connection.Conn(ctx)
	if err != nil {
		return fmt.Errorf(
			"não foi possível obter uma sessão com o MySQL: %w",
			err,
		)
	}
	defer session.Close()

	var schemaExists bool
	err = session.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.SCHEMATA
				WHERE SCHEMA_NAME = ?
			)
		`,
		schemaName,
	).Scan(&schemaExists)
	if err != nil {
		return fmt.Errorf(
			"não foi possível verificar o schema de controle: %w",
			err,
		)
	}

	if schemaExists {
		return fmt.Errorf(
			"schema de controle %q já existe",
			schemaName,
		)
	}

	createSchemaStatement := fmt.Sprintf(
		"CREATE DATABASE `%s` "+
			"CHARACTER SET utf8mb4 "+
			"COLLATE utf8mb4_0900_ai_ci",
		schemaName,
	)

	if _, err := session.ExecContext(
		ctx,
		createSchemaStatement,
	); err != nil {
		return fmt.Errorf(
			"não foi possível criar o schema de controle %q: %w",
			schemaName,
			err,
		)
	}

	if _, err := session.ExecContext(
		ctx,
		"SET NAMES utf8mb4 COLLATE utf8mb4_0900_ai_ci",
	); err != nil {
		return fmt.Errorf(
			"não foi possível configurar o charset da sessão: %w",
			err,
		)
	}

	useSchemaStatement := fmt.Sprintf(
		"USE `%s`",
		schemaName,
	)

	if _, err := session.ExecContext(
		ctx,
		useSchemaStatement,
	); err != nil {
		return fmt.Errorf(
			"não foi possível selecionar o schema de controle %q: %w",
			schemaName,
			err,
		)
	}

	startedAt := time.Now().UTC()

	if _, err := session.ExecContext(
		ctx,
		initialMigration.SQL,
	); err != nil {
		return fmt.Errorf(
			"não foi possível executar a migration %q: %w",
			initialMigration.FileName,
			err,
		)
	}

	finishedAt := time.Now().UTC()

	_, err = session.ExecContext(
		ctx,
		`
			INSERT INTO control_schema_migrations (
				version,
				name,
				checksum,
				status,
				loader_version,
				loader_commit,
				started_at_utc,
				finished_at_utc
			)
			VALUES (?, ?, ?, 'applied', ?, ?, ?, ?)
		`,
		initialMigration.Version,
		initialMigration.Name,
		initialMigration.Checksum[:],
		build.Version,
		build.Commit,
		startedAt,
		finishedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"não foi possível registrar a migration %q: %w",
			initialMigration.FileName,
			err,
		)
	}

	return nil
}
