package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"

	"address-quality/internal/model"
	"address-quality/internal/normalizer"
)

type LocationRepository struct {
	db *sql.DB
}

func NewLocationDB(dbPath string) (*LocationRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Printf("location database initialized: %s", dbPath)
	return &LocationRepository{db: db}, nil
}

func (r *LocationRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *LocationRepository) HasLocationTables(ctx context.Context) (bool, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name='location_sources'`).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *LocationRepository) FindByKode(ctx context.Context, kode string) (*model.Location, error) {
	loc := &model.Location{}

	err := r.db.QueryRowContext(ctx, `
		SELECT name, postal_code
		FROM location_codes
		WHERE kode = ? AND deleted_at IS NULL
	`, kode).Scan(&loc.SubDistrict, &loc.PostalCode)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(kode, ".")
	switch len(parts) {
	case 4:
		districtKode := parts[0] + "." + parts[1] + "." + parts[2]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND deleted_at IS NULL`, districtKode).Scan(&loc.District)
		fallthrough
	case 3:
		cityKode := parts[0] + "." + parts[1]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND deleted_at IS NULL`, cityKode).Scan(&loc.City)
		fallthrough
	case 2:
		provinceKode := parts[0]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND deleted_at IS NULL`, provinceKode).Scan(&loc.Province)
	}

	return loc, nil
}

func (r *LocationRepository) InsertLocationCode(ctx context.Context, sourceID int, kode, name string, levelID int, postalCode string) error {
	normalized := normalizer.Normalize(name)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO location_codes
			(location_source_id, kode, name, lowercase_normalized, level_id, postal_code)
		VALUES (?, ?, ?, ?, ?, ?)
	`, sourceID, kode, name, normalized, levelID, postalCode)
	return err
}

type LocationCodeRow struct {
	Kode       string
	Name       string
	LevelID    int
	PostalCode string
}

func (r *LocationRepository) DropAll(ctx context.Context) error {
	tables := []string{"location_alias", "location_codes", "location_sources", "location_levels"}
	for _, t := range tables {
		if _, err := r.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", t)); err != nil {
			return fmt.Errorf("drop %s: %w", t, err)
		}
	}
	return nil
}

func (r *LocationRepository) ExecSchema(ctx context.Context, sqlContent string) error {
	statements := strings.Split(sqlContent, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec schema: %w", err)
		}
	}
	return nil
}

func (r *LocationRepository) TruncateAll(ctx context.Context) error {
	queries := []string{
		"DELETE FROM location_alias",
		"DELETE FROM location_codes",
		"DELETE FROM location_sources",
		"DELETE FROM sqlite_sequence WHERE name IN ('location_alias', 'location_codes', 'location_sources')",
	}
	for _, q := range queries {
		if _, err := r.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("truncate: %w", err)
		}
	}
	return nil
}

func (r *LocationRepository) InsertLocationSource(ctx context.Context, code, version, name, codeDate, desc string) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO location_sources (code, version, name, code_date, desc)
		VALUES (?, ?, ?, ?, ?)
	`, code, version, name, codeDate, desc)
	if err != nil {
		return 0, fmt.Errorf("insert source: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	if id > 0 {
		return id, nil
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT id FROM location_sources WHERE code = ? AND version = ? AND deleted_at IS NULL
	`, code, version).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("lookup source: %w", err)
	}
	return id, nil
}

func (r *LocationRepository) InsertLocationCodeBatch(ctx context.Context, sourceID int64, rows []LocationCodeRow) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO location_codes (location_source_id, kode, name, lowercase_normalized, level_id, postal_code)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		normalized := normalizer.Normalize(row.Name)
		if _, err := stmt.ExecContext(ctx, sourceID, row.Kode, row.Name, normalized, row.LevelID, row.PostalCode); err != nil {
			return fmt.Errorf("insert %s: %w", row.Kode, err)
		}
	}

	return tx.Commit()
}

func (r *LocationRepository) FindByPostalCode(ctx context.Context, postalCode string) (*model.Location, error) {
	var kode string
	loc := &model.Location{}

	err := r.db.QueryRowContext(ctx, `
		SELECT kode, name, postal_code
		FROM location_codes
		WHERE postal_code = ? AND deleted_at IS NULL
		LIMIT 1
	`, postalCode).Scan(&kode, &loc.SubDistrict, &loc.PostalCode)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(kode, ".")
	switch len(parts) {
	case 4:
		districtKode := parts[0] + "." + parts[1] + "." + parts[2]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND deleted_at IS NULL`, districtKode).Scan(&loc.District)
		fallthrough
	case 3:
		cityKode := parts[0] + "." + parts[1]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND deleted_at IS NULL`, cityKode).Scan(&loc.City)
		fallthrough
	case 2:
		provinceKode := parts[0]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND deleted_at IS NULL`, provinceKode).Scan(&loc.Province)
	}

	return loc, nil
}
