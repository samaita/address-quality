package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"

	"address-quality/internal/logger"
	"address-quality/internal/model"
	"address-quality/internal/normalizer"
)

type LocationRepository struct {
	db *sql.DB
}

func NewLocationDB(dbPath string, maxOpenConns int) (*LocationRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	logger.Info().Str("db_path", dbPath).Msg("location database initialized")
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

func (r *LocationRepository) FindSourceByCode(ctx context.Context, code string) (int64, string, error) {
	var id int64
	var version string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, version FROM location_sources
		WHERE code = ? AND deleted_at IS NULL
		ORDER BY id DESC
		LIMIT 1
	`, code).Scan(&id, &version)
	if err != nil {
		return 0, "", fmt.Errorf("find source by code %s: %w", code, err)
	}
	return id, version, nil
}

func (r *LocationRepository) FindByKode(ctx context.Context, kode string, sourceID int64) (*model.Location, error) {
	loc := &model.Location{}

	err := r.db.QueryRowContext(ctx, `
		SELECT name, postal_code
		FROM location_codes
		WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL
	`, kode, sourceID).Scan(&loc.SubDistrict, &loc.PostalCode)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(kode, ".")
	switch len(parts) {
	case 4:
		districtKode := parts[0] + "." + parts[1] + "." + parts[2]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL`, districtKode, sourceID).Scan(&loc.District)
		fallthrough
	case 3:
		cityKode := parts[0] + "." + parts[1]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL`, cityKode, sourceID).Scan(&loc.City)
		fallthrough
	case 2:
		provinceKode := parts[0]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL`, provinceKode, sourceID).Scan(&loc.Province)
	}

	return loc, nil
}

type ProvinceRow struct {
	SourceID            int64
	ProvinceID          int64
	Name                string
	LowercaseNormalized string
}

func (r *LocationRepository) FindAllProvinces(ctx context.Context) ([]ProvinceRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT location_source_id, id, name, lowercase_normalized
		FROM location_codes
		WHERE level_id = 2 AND deleted_at IS NULL
	`)
	if err != nil {
		return nil, fmt.Errorf("find all provinces: %w", err)
	}
	defer rows.Close()

	var provinces []ProvinceRow
	for rows.Next() {
		var p ProvinceRow
		if err := rows.Scan(&p.SourceID, &p.ProvinceID, &p.Name, &p.LowercaseNormalized); err != nil {
			return nil, fmt.Errorf("scan province: %w", err)
		}
		provinces = append(provinces, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows provinces: %w", err)
	}
	return provinces, nil
}

func (r *LocationRepository) FindProvincesBySourceID(ctx context.Context, sourceID int64) ([]ProvinceRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT location_source_id, id, name, lowercase_normalized
		FROM location_codes
		WHERE level_id = 2 AND location_source_id = ? AND deleted_at IS NULL
	`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("find provinces: %w", err)
	}
	defer rows.Close()

	var provinces []ProvinceRow
	for rows.Next() {
		var p ProvinceRow
		if err := rows.Scan(&p.SourceID, &p.ProvinceID, &p.Name, &p.LowercaseNormalized); err != nil {
			return nil, fmt.Errorf("scan province: %w", err)
		}
		provinces = append(provinces, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows provinces: %w", err)
	}
	return provinces, nil
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
	tables := []string{"location_hierarchy", "location_alias", "location_codes", "location_sources", "location_levels"}
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
		lines := strings.Split(stmt, "\n")
		start := 0
		for start < len(lines) {
			trimmed := strings.TrimSpace(lines[start])
			if trimmed == "" || strings.HasPrefix(trimmed, "--") {
				start++
			} else {
				break
			}
		}
		stmt = strings.TrimSpace(strings.Join(lines[start:], "\n"))
		if stmt == "" {
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
		"DELETE FROM location_hierarchy",
		"DELETE FROM location_alias",
		"DELETE FROM location_codes",
		"DELETE FROM location_sources",
		"DELETE FROM sqlite_sequence WHERE name IN ('location_hierarchy', 'location_alias', 'location_codes', 'location_sources')",
	}
	for _, q := range queries {
		if _, err := r.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("truncate: %w", err)
		}
	}
	return nil
}

func (r *LocationRepository) RebuildLocationHierarchy(ctx context.Context, sourceID int64) error {
	logger.Info().Msg("building location hierarchy...")

	rows, err := r.db.QueryContext(ctx, `
		SELECT kode, id FROM location_codes
		WHERE location_source_id = ? AND level_id IN (2, 3, 4) AND deleted_at IS NULL
	`, sourceID)
	if err != nil {
		return fmt.Errorf("query parents: %w", err)
	}
	defer rows.Close()

	kodeToID := make(map[string]int64)
	for rows.Next() {
		var kode string
		var id int64
		if err := rows.Scan(&kode, &id); err != nil {
			return fmt.Errorf("scan parent: %w", err)
		}
		kodeToID[kode] = id
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows parents: %w", err)
	}
	logger.Info().Int("count", len(kodeToID)).Msg("loaded parent kode->id mappings")

	subRows, err := r.db.QueryContext(ctx, `
		SELECT id, kode FROM location_codes
		WHERE location_source_id = ? AND level_id = 5 AND deleted_at IS NULL
	`, sourceID)
	if err != nil {
		return fmt.Errorf("query subdistricts: %w", err)
	}
	defer subRows.Close()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO location_hierarchy (location_source_id, province_id, city_id, district_id, subdistrict_id)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	var count int
	for subRows.Next() {
		var id int64
		var kode string
		if err := subRows.Scan(&id, &kode); err != nil {
			return fmt.Errorf("scan subdistrict: %w", err)
		}

		parts := strings.Split(kode, ".")
		if len(parts) != 4 {
			continue
		}

		provinceKode := parts[0]
		cityKode := parts[0] + "." + parts[1]
		districtKode := parts[0] + "." + parts[1] + "." + parts[2]

		provinceID, ok := kodeToID[provinceKode]
		if !ok {
			logger.Warn().Str("kode", kode).Str("parent_province", provinceKode).Msg("skipping: parent province not found")
			continue
		}
		cityID, ok := kodeToID[cityKode]
		if !ok {
			logger.Warn().Str("kode", kode).Str("parent_city", cityKode).Msg("skipping: parent city not found")
			continue
		}
		districtID, ok := kodeToID[districtKode]
		if !ok {
			logger.Warn().Str("kode", kode).Str("parent_district", districtKode).Msg("skipping: parent district not found")
			continue
		}

		if _, err := stmt.ExecContext(ctx, sourceID, provinceID, cityID, districtID, id); err != nil {
			return fmt.Errorf("insert hierarchy %s: %w", kode, err)
		}
		count++
	}

	if err := subRows.Err(); err != nil {
		return fmt.Errorf("rows subdistricts: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	logger.Info().Int("count", count).Msg("inserted hierarchy rows")
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

func (r *LocationRepository) FindByPostalCode(ctx context.Context, postalCode string, sourceID int64) (*model.Location, error) {
	var kode string
	loc := &model.Location{}

	err := r.db.QueryRowContext(ctx, `
		SELECT kode, name, postal_code
		FROM location_codes
		WHERE postal_code = ? AND location_source_id = ? AND deleted_at IS NULL
		LIMIT 1
	`, postalCode, sourceID).Scan(&kode, &loc.SubDistrict, &loc.PostalCode)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(kode, ".")
	switch len(parts) {
	case 4:
		districtKode := parts[0] + "." + parts[1] + "." + parts[2]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL`, districtKode, sourceID).Scan(&loc.District)
		fallthrough
	case 3:
		cityKode := parts[0] + "." + parts[1]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL`, cityKode, sourceID).Scan(&loc.City)
		fallthrough
	case 2:
		provinceKode := parts[0]
		r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL`, provinceKode, sourceID).Scan(&loc.Province)
	}

	return loc, nil
}
