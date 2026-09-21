// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSQLRowsCopySource(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE sample (id INTEGER, name TEXT, note TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sample VALUES (1, 'Aceh', NULL), (2, 'Bali', 'island')`); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT id, name, note FROM sample ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	source := newSQLRowsCopySource(rows, 3)
	var got [][]any
	for source.Next() {
		values, err := source.Values()
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, append([]any(nil), values...))
	}
	if err := source.Err(); err != nil {
		t.Fatal(err)
	}
	want := [][]any{{int64(1), "Aceh", nil}, {int64(2), "Bali", "island"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("copied rows = %#v, want %#v", got, want)
	}
}

func TestReplaceFromSQLiteIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not set")
	}

	ctx := context.Background()
	sqliteRepo, err := NewLocationDB(filepath.Join(t.TempDir(), "location.db"), 1)
	if err != nil {
		t.Fatal(err)
	}
	sqliteSchema, err := os.ReadFile("../../db/location.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := sqliteRepo.ExecSchema(ctx, string(sqliteSchema)); err != nil {
		t.Fatal(err)
	}
	if _, err := sqliteRepo.db.ExecContext(ctx, `
		INSERT INTO location_sources (id, code, version, name) VALUES (7, 'test', '1', 'Test');
		INSERT INTO location_codes (id, location_source_id, kode, name, lowercase_normalized, level_id)
		VALUES (42, 7, '11', 'Aceh', 'aceh', 2);
		INSERT INTO location_alias (id, location_id, alias) VALUES (3, 42, 'NAD');
		INSERT INTO location_city_priority (id, location_source_id, lowercase_normalized, city_type)
		VALUES (9, 7, 'aceh', 'KOTA');
	`); err != nil {
		t.Fatal(err)
	}

	pgRepo, err := NewPostgresDB(dsn, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer pgRepo.DropAll(ctx)
	pgSchema, err := os.ReadFile("../../db/location_postgres.sql")
	if err != nil {
		t.Fatal(err)
	}
	pgIndexes, err := os.ReadFile("../../db/location_postgres_indexes.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := pgRepo.ReplaceFromSQLite(ctx, sqliteRepo, string(pgSchema), string(pgIndexes)); err != nil {
		t.Fatal(err)
	}

	for table, want := range map[string]int64{
		"location_levels": 5, "location_sources": 1, "location_codes": 1,
		"location_alias": 1, "location_hierarchy": 0, "location_city_priority": 1,
	} {
		var got int64
		if err := pgRepo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("%s count = %d, want %d", table, got, want)
		}
	}

	var id int64
	if err := pgRepo.db.QueryRowContext(ctx, `
		INSERT INTO location_sources (code, version, name)
		VALUES ('sequence-test', '1', 'Sequence Test') RETURNING id
	`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if id <= 7 {
		t.Fatalf("next location_sources id = %d, want > 7", id)
	}

	var indexCount int
	if err := pgRepo.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pg_indexes
		WHERE tablename = 'location_codes'
		  AND indexname IN ('location_codes_normalized_idx', 'location_codes_normalized_trgm_idx')
	`).Scan(&indexCount); err != nil {
		t.Fatal(err)
	}
	if indexCount != 2 {
		t.Fatalf("location search index count = %d, want 2", indexCount)
	}

	var similarity float64
	if err := pgRepo.db.QueryRowContext(ctx, `SELECT similarity('aceh', 'ace')`).Scan(&similarity); err != nil {
		t.Fatal(err)
	}
}
