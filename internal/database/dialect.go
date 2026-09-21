// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// sqlDB wraps *sql.DB and rewrites `?` placeholders to `$n` for Postgres, so
// repository queries stay dialect-neutral.
type sqlDB struct {
	*sql.DB
	pg bool
}

func (d *sqlDB) rebind(query string) string {
	if !d.pg {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 8)
	n := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			n++
			fmt.Fprintf(&b, "$%d", n)
			continue
		}
		b.WriteByte(query[i])
	}
	return b.String()
}

func (d *sqlDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.DB.QueryContext(ctx, d.rebind(query), args...)
}

func (d *sqlDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.DB.QueryRowContext(ctx, d.rebind(query), args...)
}

func (d *sqlDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.DB.ExecContext(ctx, d.rebind(query), args...)
}

func (d *sqlDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return d.DB.PrepareContext(ctx, d.rebind(query))
}

// stripSQLComments removes full-line SQL comments so statement splitting on
// ";" is not confused by semicolons inside comments.
func stripSQLComments(sqlContent string) string {
	var b strings.Builder
	for _, line := range strings.Split(sqlContent, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}
