// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type postgresCopyTable struct {
	name    string
	columns []string
}

var postgresCopyTables = []postgresCopyTable{
	{name: "location_levels", columns: []string{"id", "code", "name"}},
	{name: "location_sources", columns: []string{"id", "code", "version", "name", "code_date", "desc", "created_at", "updated_at", "deleted_at"}},
	{name: "location_codes", columns: []string{"id", "location_source_id", "kode", "name", "lowercase_normalized", "level_id", "postal_code", "created_at", "updated_at", "deleted_at"}},
	{name: "location_alias", columns: []string{"id", "location_id", "alias", "created_at", "updated_at", "deleted_at"}},
	{name: "location_hierarchy", columns: []string{"id", "location_source_id", "province_id", "city_id", "district_id", "subdistrict_id", "created_at", "updated_at", "deleted_at"}},
	{name: "location_city_priority", columns: []string{"id", "location_source_id", "lowercase_normalized", "city_type", "created_at", "updated_at", "deleted_at"}},
}

// ReplaceFromSQLite atomically replaces all Postgres location tables with an
// exact snapshot of the SQLite source of truth. Schema creation, COPY, sequence
// repair, validation, and index creation share one Postgres transaction, so a
// failed refresh leaves the previous Postgres snapshot visible.
func (r *LocationRepository) ReplaceFromSQLite(ctx context.Context, source *LocationRepository, schema, indexes string) error {
	if r == nil || r.db == nil || !r.db.pg {
		return fmt.Errorf("replace from sqlite: destination is not postgres")
	}
	if source == nil || source.db == nil || source.db.pg {
		return fmt.Errorf("replace from sqlite: source is not sqlite")
	}

	sourceTx, err := source.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return logDBErr(ctx, "postgres_copy_begin_sqlite", "", err)
	}
	defer sourceTx.Rollback()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return logDBErr(ctx, "postgres_copy_acquire", "", err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn any) error {
		stdlibConn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return fmt.Errorf("unexpected postgres driver connection %T", driverConn)
		}

		pgTx, err := stdlibConn.Conn().Begin(ctx)
		if err != nil {
			return err
		}
		defer pgTx.Rollback(ctx)

		if _, err := pgTx.Exec(ctx, `
			DROP TABLE IF EXISTS
				location_hierarchy,
				location_alias,
				location_city_priority,
				location_codes,
				location_sources,
				location_levels
			CASCADE
		`); err != nil {
			return fmt.Errorf("drop postgres location tables: %w", err)
		}
		if err := execPGXSchema(ctx, pgTx, schema); err != nil {
			return fmt.Errorf("create postgres location schema: %w", err)
		}

		for _, table := range postgresCopyTables {
			if err := copySQLiteTable(ctx, sourceTx, pgTx, table); err != nil {
				return err
			}
		}
		if err := resetPostgresSequences(ctx, pgTx); err != nil {
			return err
		}
		if err := execPGXSchema(ctx, pgTx, indexes); err != nil {
			return fmt.Errorf("create postgres location indexes: %w", err)
		}

		return pgTx.Commit(ctx)
	})
	return logDBErr(ctx, "postgres_copy_snapshot", "", err)
}

func execPGXSchema(ctx context.Context, tx pgx.Tx, sqlContent string) error {
	for _, statement := range strings.Split(stripSQLComments(sqlContent), ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := tx.Exec(ctx, statement); err != nil {
			return fmt.Errorf("exec %q: %w", statement, err)
		}
	}
	return nil
}

func copySQLiteTable(ctx context.Context, sourceTx *sql.Tx, targetTx pgx.Tx, table postgresCopyTable) error {
	quotedColumns := make([]string, len(table.columns))
	for i, column := range table.columns {
		quotedColumns[i] = `"` + strings.ReplaceAll(column, `"`, `""`) + `"`
	}
	query := fmt.Sprintf(`SELECT %s FROM "%s" ORDER BY id`, strings.Join(quotedColumns, ", "), table.name)
	rows, err := sourceTx.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("read sqlite table %s: %w", table.name, err)
	}
	defer rows.Close()

	source := newSQLRowsCopySource(rows, len(table.columns))
	copied, err := targetTx.CopyFrom(ctx, pgx.Identifier{table.name}, table.columns, source)
	if err != nil {
		return fmt.Errorf("copy table %s: %w", table.name, err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close sqlite rows for %s: %w", table.name, err)
	}
	if err := source.Err(); err != nil {
		return fmt.Errorf("read sqlite rows for %s: %w", table.name, err)
	}

	var stored int64
	if err := targetTx.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, table.name)).Scan(&stored); err != nil {
		return fmt.Errorf("count copied table %s: %w", table.name, err)
	}
	if copied != stored {
		return fmt.Errorf("validate copied table %s: COPY reported %d rows but postgres contains %d", table.name, copied, stored)
	}
	return nil
}

func resetPostgresSequences(ctx context.Context, tx pgx.Tx) error {
	for _, table := range postgresCopyTables {
		query := fmt.Sprintf(`
			SELECT setval(
				pg_get_serial_sequence('%s', 'id'),
				COALESCE(MAX(id), 1),
				COUNT(*) > 0
			)
			FROM "%s"
		`, table.name, table.name)
		if _, err := tx.Exec(ctx, query); err != nil {
			return fmt.Errorf("reset postgres sequence for %s: %w", table.name, err)
		}
	}
	return nil
}

type sqlRowsCopySource struct {
	rows    *sql.Rows
	values  []any
	targets []any
	err     error
}

func newSQLRowsCopySource(rows *sql.Rows, columnCount int) *sqlRowsCopySource {
	values := make([]any, columnCount)
	targets := make([]any, columnCount)
	for i := range values {
		targets[i] = &values[i]
	}
	return &sqlRowsCopySource{rows: rows, values: values, targets: targets}
}

func (s *sqlRowsCopySource) Next() bool {
	if !s.rows.Next() {
		return false
	}
	s.err = s.rows.Scan(s.targets...)
	return s.err == nil
}

func (s *sqlRowsCopySource) Values() ([]any, error) {
	return s.values, s.err
}

func (s *sqlRowsCopySource) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.rows.Err()
}
