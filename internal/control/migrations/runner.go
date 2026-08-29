package migrations

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/danielsantello/go-cnpj-loader/internal/buildinfo"
	"github.com/go-sql-driver/mysql"
)

var schemaNamePattern = regexp.MustCompile(
	`^[a-z][a-z0-9_]{0,63}$`,
)

func Migrate(
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

	if !schemaExists {
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

	pendingMigrations := catalog

	if schemaExists {
		pendingMigrations, err = findPendingMigrations(
			ctx,
			session,
			catalog,
		)
		if err != nil {
			return err
		}
	}

	for _, migration := range pendingMigrations {
		if err := applyMigration(
			ctx,
			session,
			migration,
			build,
		); err != nil {
			return err
		}
	}

	return nil
}

func findPendingMigrations(
	ctx context.Context,
	session *sql.Conn,
	catalog []Migration,
) ([]Migration, error) {
	rows, err := session.QueryContext(
		ctx,
		`
			SELECT
				version,
				name,
				checksum,
				status
			FROM control_schema_migrations
			ORDER BY version
		`,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível consultar o histórico de migrations: %w",
			err,
		)
	}
	defer rows.Close()

	appliedCount := 0

	for rows.Next() {
		if appliedCount >= len(catalog) {
			return nil, fmt.Errorf(
				"histórico possui migrations ausentes no catálogo incorporado",
			)
		}

		var (
			version  uint32
			name     string
			checksum []byte
			status   string
		)

		if err := rows.Scan(
			&version,
			&name,
			&checksum,
			&status,
		); err != nil {
			return nil, fmt.Errorf(
				"não foi possível ler o histórico de migrations: %w",
				err,
			)
		}

		expected := catalog[appliedCount]

		if version != expected.Version {
			return nil, fmt.Errorf(
				"histórico de migrations possui versão %d; esperava %d",
				version,
				expected.Version,
			)
		}

		if name != expected.Name {
			return nil, fmt.Errorf(
				"migration %d possui nome %q; esperava %q",
				version,
				name,
				expected.Name,
			)
		}

		if !bytes.Equal(checksum, expected.Checksum[:]) {
			return nil, fmt.Errorf(
				"migration %d possui checksum diferente do catálogo incorporado",
				version,
			)
		}

		if status != "applied" {
			return nil, fmt.Errorf(
				"migration %d possui status %q; esperava %q",
				version,
				status,
				"applied",
			)
		}

		appliedCount++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"não foi possível percorrer o histórico de migrations: %w",
			err,
		)
	}

	if appliedCount == 0 {
		return nil, fmt.Errorf(
			"schema de controle existente não possui histórico de migrations",
		)
	}

	return catalog[appliedCount:], nil
}

func applyMigration(
	ctx context.Context,
	session *sql.Conn,
	migration Migration,
	build buildinfo.Info,
) error {
	if migration.Version == 1 {
		return applyInitialMigration(
			ctx,
			session,
			migration,
			build,
		)
	}

	return applyTrackedMigration(
		ctx,
		session,
		migration,
		build,
	)
}

func applyInitialMigration(
	ctx context.Context,
	session *sql.Conn,
	migration Migration,
	build buildinfo.Info,
) error {
	startedAt := time.Now().UTC()

	if _, err := session.ExecContext(
		ctx,
		migration.SQL,
	); err != nil {
		return fmt.Errorf(
			"não foi possível executar a migration %q: %w",
			migration.FileName,
			err,
		)
	}

	finishedAt := time.Now().UTC()

	_, err := session.ExecContext(
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
		migration.Version,
		migration.Name,
		migration.Checksum[:],
		build.Version,
		build.Commit,
		startedAt,
		finishedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"não foi possível registrar a migration %q: %w",
			migration.FileName,
			err,
		)
	}

	return nil
}

func applyTrackedMigration(
	ctx context.Context,
	session *sql.Conn,
	migration Migration,
	build buildinfo.Info,
) error {
	startedAt := time.Now().UTC()

	_, err := session.ExecContext(
		ctx,
		`
			INSERT INTO control_schema_migrations (
				version,
				name,
				checksum,
				status,
				loader_version,
				loader_commit,
				started_at_utc
			)
			VALUES (?, ?, ?, 'applying', ?, ?, ?)
		`,
		migration.Version,
		migration.Name,
		migration.Checksum[:],
		build.Version,
		build.Commit,
		startedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"não foi possível iniciar o registro da migration %q: %w",
			migration.FileName,
			err,
		)
	}

	if _, err := session.ExecContext(
		ctx,
		migration.SQL,
	); err != nil {
		finishedAt := time.Now().UTC()
		mysqlErrorCode, sqlState := mysqlErrorDetails(err)

		_, recordError := session.ExecContext(
			ctx,
			`
				UPDATE control_schema_migrations
				SET
					status = 'failed',
					finished_at_utc = ?,
					mysql_error_code = ?,
					sql_state = ?,
					error_message = ?
				WHERE version = ?
					AND status = 'applying'
			`,
			finishedAt,
			mysqlErrorCode,
			sqlState,
			err.Error(),
			migration.Version,
		)

		migrationError := fmt.Errorf(
			"não foi possível executar a migration %q: %w",
			migration.FileName,
			err,
		)

		if recordError != nil {
			return errors.Join(
				migrationError,
				fmt.Errorf(
					"não foi possível registrar a falha da migration %q: %w",
					migration.FileName,
					recordError,
				),
			)
		}

		return migrationError
	}

	finishedAt := time.Now().UTC()

	_, err = session.ExecContext(
		ctx,
		`
			UPDATE control_schema_migrations
			SET
				status = 'applied',
				finished_at_utc = ?
			WHERE version = ?
				AND status = 'applying'
		`,
		finishedAt,
		migration.Version,
	)
	if err != nil {
		return fmt.Errorf(
			"não foi possível concluir o registro da migration %q: %w",
			migration.FileName,
			err,
		)
	}

	return nil
}

func mysqlErrorDetails(err error) (sql.NullInt64, sql.NullString) {
	var mysqlError *mysql.MySQLError

	if !errors.As(err, &mysqlError) {
		return sql.NullInt64{}, sql.NullString{}
	}

	return sql.NullInt64{
			Int64: int64(mysqlError.Number),
			Valid: true,
		},
		sql.NullString{
			String: string(mysqlError.SQLState[:]),
			Valid:  true,
		}
}
