// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

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
		return nil, logDBErr(context.Background(), "open", dbPath, err)
	}

	db.SetMaxOpenConns(maxOpenConns)

	if err = db.Ping(); err != nil {
		return nil, logDBErr(context.Background(), "ping", dbPath, err)
	}

	logger.Info().Str("db_path", dbPath).Msg("location database initialized")
	return &LocationRepository{db: db}, nil
}

func (r *LocationRepository) Ping(ctx context.Context) error {
	return logDBErr(ctx, "ping", "", r.db.PingContext(ctx))
}

func (r *LocationRepository) HasLocationTables(ctx context.Context) (bool, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name='location_sources'`).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, logDBErr(ctx, "has_location_tables", "", err)
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
		return 0, "", logDBErr(ctx, "find_source_by_code", code, fmt.Errorf("find source by code %s: %w", code, err))
	}
	return id, version, nil
}

type ProvinceRow struct {
	ID                  int64
	SourceID            int64
	Kode                string
	Name                string
	LowercaseNormalized string
}

func (r *LocationRepository) FindAllProvinces(ctx context.Context) ([]ProvinceRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, location_source_id, kode, name, lowercase_normalized
		FROM location_codes
		WHERE level_id = 2 AND deleted_at IS NULL
	`)
	if err != nil {
		return nil, logDBErr(ctx, "find_all_provinces", "", fmt.Errorf("find all provinces: %w", err))
	}
	defer rows.Close()

	var provinces []ProvinceRow
	for rows.Next() {
		var p ProvinceRow
		if err := rows.Scan(&p.ID, &p.SourceID, &p.Kode, &p.Name, &p.LowercaseNormalized); err != nil {
			return nil, logDBErr(ctx, "find_all_provinces_scan", "", fmt.Errorf("scan province: %w", err))
		}
		provinces = append(provinces, p)
	}
	if err := rows.Err(); err != nil {
		return nil, logDBErr(ctx, "find_all_provinces_rows", "", fmt.Errorf("rows provinces: %w", err))
	}
	return provinces, nil
}

type CityRow struct {
	ID                  int64
	SourceID            int64
	Kode                string
	Name                string
	LowercaseNormalized string
	PostalCode          string
}

func (r *LocationRepository) FindAllCities(ctx context.Context) ([]CityRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, location_source_id, kode, name, lowercase_normalized, COALESCE(postal_code, '')
		FROM location_codes
		WHERE level_id = 3 AND deleted_at IS NULL
	`)
	if err != nil {
		return nil, logDBErr(ctx, "find_all_cities", "", fmt.Errorf("find all cities: %w", err))
	}
	defer rows.Close()

	var cities []CityRow
	for rows.Next() {
		var c CityRow
		if err := rows.Scan(&c.ID, &c.SourceID, &c.Kode, &c.Name, &c.LowercaseNormalized, &c.PostalCode); err != nil {
			return nil, logDBErr(ctx, "find_all_cities_scan", "", fmt.Errorf("scan city: %w", err))
		}
		cities = append(cities, c)
	}
	if err := rows.Err(); err != nil {
		return nil, logDBErr(ctx, "find_all_cities_rows", "", fmt.Errorf("rows cities: %w", err))
	}
	return cities, nil
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
			return logDBErr(ctx, "drop_all", t, fmt.Errorf("drop %s: %w", t, err))
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
			return logDBErr(ctx, "exec_schema", stmt, fmt.Errorf("exec schema: %w", err))
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
			return logDBErr(ctx, "truncate_all", q, fmt.Errorf("truncate: %w", err))
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
		return logDBErr(ctx, "rebuild_hierarchy_query_parents", sourceID, fmt.Errorf("query parents: %w", err))
	}
	defer rows.Close()

	kodeToID := make(map[string]int64)
	for rows.Next() {
		var kode string
		var id int64
		if err := rows.Scan(&kode, &id); err != nil {
			return logDBErr(ctx, "rebuild_hierarchy_scan_parent", sourceID, fmt.Errorf("scan parent: %w", err))
		}
		kodeToID[kode] = id
	}
	if err := rows.Err(); err != nil {
		return logDBErr(ctx, "rebuild_hierarchy_rows_parents", sourceID, fmt.Errorf("rows parents: %w", err))
	}
	logger.Info().Int("count", len(kodeToID)).Msg("loaded parent kode->id mappings")

	subRows, err := r.db.QueryContext(ctx, `
		SELECT id, kode FROM location_codes
		WHERE location_source_id = ? AND level_id = 5 AND deleted_at IS NULL
	`, sourceID)
	if err != nil {
		return logDBErr(ctx, "rebuild_hierarchy_query_subdistricts", sourceID, fmt.Errorf("query subdistricts: %w", err))
	}
	defer subRows.Close()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return logDBErr(ctx, "rebuild_hierarchy_begin_tx", sourceID, fmt.Errorf("begin tx: %w", err))
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO location_hierarchy (location_source_id, province_id, city_id, district_id, subdistrict_id)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return logDBErr(ctx, "rebuild_hierarchy_prepare", sourceID, fmt.Errorf("prepare: %w", err))
	}
	defer stmt.Close()

	var count int
	for subRows.Next() {
		var id int64
		var kode string
		if err := subRows.Scan(&id, &kode); err != nil {
			return logDBErr(ctx, "rebuild_hierarchy_scan_subdistrict", kode, fmt.Errorf("scan subdistrict: %w", err))
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
			return logDBErr(ctx, "rebuild_hierarchy_insert", kode, fmt.Errorf("insert hierarchy %s: %w", kode, err))
		}
		count++
	}

	if err := subRows.Err(); err != nil {
		return logDBErr(ctx, "rebuild_hierarchy_rows_subdistricts", sourceID, fmt.Errorf("rows subdistricts: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return logDBErr(ctx, "rebuild_hierarchy_commit", sourceID, fmt.Errorf("commit: %w", err))
	}

	logger.Info().Int("count", count).Msg("inserted hierarchy rows")
	return nil
}

func (r *LocationRepository) RebuildNormalized(ctx context.Context) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name FROM location_codes WHERE deleted_at IS NULL
	`)
	if err != nil {
		return logDBErr(ctx, "rebuild_normalized_query", "", fmt.Errorf("rebuild normalized query: %w", err))
	}
	defer rows.Close()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return logDBErr(ctx, "rebuild_normalized_begin_tx", "", fmt.Errorf("begin tx: %w", err))
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE location_codes SET lowercase_normalized = ?, updated_at = datetime('now') WHERE id = ?
	`)
	if err != nil {
		return logDBErr(ctx, "rebuild_normalized_prepare", "", fmt.Errorf("prepare: %w", err))
	}
	defer stmt.Close()

	var count int
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return logDBErr(ctx, "rebuild_normalized_scan", id, fmt.Errorf("scan: %w", err))
		}
		normalized := normalizer.Normalize(name)
		if _, err := stmt.ExecContext(ctx, normalized, id); err != nil {
			return logDBErr(ctx, "rebuild_normalized_update", id, fmt.Errorf("update %d: %w", id, err))
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return logDBErr(ctx, "rebuild_normalized_rows", "", fmt.Errorf("rows: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return logDBErr(ctx, "rebuild_normalized_commit", "", fmt.Errorf("commit: %w", err))
	}

	logger.Info().Int("count", count).Msg("rebuilt lowercase_normalized")
	return nil
}

func (r *LocationRepository) InsertLocationSource(ctx context.Context, code, version, name, codeDate, desc string) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO location_sources (code, version, name, code_date, desc)
		VALUES (?, ?, ?, ?, ?)
	`, code, version, name, codeDate, desc)
	if err != nil {
		return 0, logDBErr(ctx, "insert_source", map[string]any{"code": code, "version": version, "name": name}, fmt.Errorf("insert source: %w", err))
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, logDBErr(ctx, "insert_source_last_insert_id", code, fmt.Errorf("last insert id: %w", err))
	}

	if id > 0 {
		return id, nil
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT id FROM location_sources WHERE code = ? AND version = ? AND deleted_at IS NULL
	`, code, version).Scan(&id)
	if err != nil {
		return 0, logDBErr(ctx, "insert_source_lookup", map[string]any{"code": code, "version": version}, fmt.Errorf("lookup source: %w", err))
	}
	return id, nil
}

func (r *LocationRepository) InsertLocationCodeBatch(ctx context.Context, sourceID int64, rows []LocationCodeRow) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return logDBErr(ctx, "insert_code_batch_begin_tx", sourceID, fmt.Errorf("begin tx: %w", err))
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO location_codes (location_source_id, kode, name, lowercase_normalized, level_id, postal_code)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return logDBErr(ctx, "insert_code_batch_prepare", sourceID, fmt.Errorf("prepare: %w", err))
	}
	defer stmt.Close()

	for _, row := range rows {
		normalized := normalizer.Normalize(row.Name)
		if _, err := stmt.ExecContext(ctx, sourceID, row.Kode, row.Name, normalized, row.LevelID, row.PostalCode); err != nil {
			return logDBErr(ctx, "insert_code_batch", map[string]any{"source_id": sourceID, "kode": row.Kode}, fmt.Errorf("insert %s: %w", row.Kode, err))
		}
	}

	return logDBErr(ctx, "insert_code_batch_commit", sourceID, tx.Commit())
}

func (r *LocationRepository) FindByPostalCode(ctx context.Context, postalCode string, sourceID int64) ([]model.Location, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT kode, name, postal_code
		FROM location_codes
		WHERE postal_code = ? AND location_source_id = ? AND deleted_at IS NULL
	`, postalCode, sourceID)
	if err != nil {
		return nil, logDBErr(ctx, "find_by_postal_code", map[string]any{"postal_code": postalCode, "source_id": sourceID}, err)
	}
	defer rows.Close()

	var results []model.Location
	for rows.Next() {
		var kode string
		loc := model.Location{}
		if err := rows.Scan(&kode, &loc.SubDistrict, &loc.PostalCode); err != nil {
			return nil, logDBErr(ctx, "find_by_postal_code_scan", map[string]any{"postal_code": postalCode, "source_id": sourceID}, err)
		}

		parts := strings.Split(kode, ".")
		switch len(parts) {
		case 4:
			districtKode := parts[0] + "." + parts[1] + "." + parts[2]
			if err := r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL`, districtKode, sourceID).Scan(&loc.District); err != nil {
				logDBErr(ctx, "find_by_postal_code_lookup_district", map[string]any{"kode": districtKode, "source_id": sourceID}, err)
			}
			fallthrough
		case 3:
			cityKode := parts[0] + "." + parts[1]
			if err := r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL`, cityKode, sourceID).Scan(&loc.City); err != nil {
				logDBErr(ctx, "find_by_postal_code_lookup_city", map[string]any{"kode": cityKode, "source_id": sourceID}, err)
			}
			fallthrough
		case 2:
			provinceKode := parts[0]
			if err := r.db.QueryRowContext(ctx, `SELECT name FROM location_codes WHERE kode = ? AND location_source_id = ? AND deleted_at IS NULL`, provinceKode, sourceID).Scan(&loc.Province); err != nil {
				logDBErr(ctx, "find_by_postal_code_lookup_province", map[string]any{"kode": provinceKode, "source_id": sourceID}, err)
			}
		}

		results = append(results, loc)
	}
	if err := rows.Err(); err != nil {
		return nil, logDBErr(ctx, "find_by_postal_code_rows", map[string]any{"postal_code": postalCode, "source_id": sourceID}, err)
	}

	return results, nil
}

type DistrictRow struct {
	ID                  int64
	SourceID            int64
	Kode                string
	Name                string
	LowercaseNormalized string
}

func (r *LocationRepository) FindAllDistricts(ctx context.Context, sourceID int64) ([]DistrictRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, location_source_id, kode, name, lowercase_normalized
		FROM location_codes
		WHERE level_id = 4 AND location_source_id = ? AND deleted_at IS NULL
	`, sourceID)
	if err != nil {
		return nil, logDBErr(ctx, "find_all_districts", sourceID, fmt.Errorf("find all districts: %w", err))
	}
	defer rows.Close()

	var districts []DistrictRow
	for rows.Next() {
		var d DistrictRow
		if err := rows.Scan(&d.ID, &d.SourceID, &d.Kode, &d.Name, &d.LowercaseNormalized); err != nil {
			return nil, logDBErr(ctx, "find_all_districts_scan", sourceID, fmt.Errorf("scan district: %w", err))
		}
		districts = append(districts, d)
	}
	if err := rows.Err(); err != nil {
		return nil, logDBErr(ctx, "find_all_districts_rows", sourceID, fmt.Errorf("rows districts: %w", err))
	}
	return districts, nil
}

type SubDistrictRow struct {
	ID                  int64
	SourceID            int64
	Kode                string
	Name                string
	LowercaseNormalized string
	PostalCode          string
}

func (r *LocationRepository) FindAllSubDistricts(ctx context.Context, sourceID int64) ([]SubDistrictRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, location_source_id, kode, name, lowercase_normalized, COALESCE(postal_code, '')
		FROM location_codes
		WHERE level_id = 5 AND location_source_id = ? AND deleted_at IS NULL
	`, sourceID)
	if err != nil {
		return nil, logDBErr(ctx, "find_all_subdistricts", sourceID, fmt.Errorf("find all subdistricts: %w", err))
	}
	defer rows.Close()

	var subdistricts []SubDistrictRow
	for rows.Next() {
		var s SubDistrictRow
		if err := rows.Scan(&s.ID, &s.SourceID, &s.Kode, &s.Name, &s.LowercaseNormalized, &s.PostalCode); err != nil {
			return nil, logDBErr(ctx, "find_all_subdistricts_scan", sourceID, fmt.Errorf("scan subdistrict: %w", err))
		}
		subdistricts = append(subdistricts, s)
	}
	if err := rows.Err(); err != nil {
		return nil, logDBErr(ctx, "find_all_subdistricts_rows", sourceID, fmt.Errorf("rows subdistricts: %w", err))
	}
	return subdistricts, nil
}

type HierarchyRow struct {
	ProvinceID    int64
	CityID        int64
	DistrictID    int64
	SubDistrictID int64
}

type HierarchyMap struct {
	CityToProvince    map[int64]int64
	DistrictToCity    map[int64]int64
	SubDistrictToDist map[int64]int64
}

func (r *LocationRepository) LoadFullHierarchy(ctx context.Context, sourceID int64) (*HierarchyMap, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT province_id, city_id, district_id, subdistrict_id
		FROM location_hierarchy
		WHERE location_source_id = ? AND deleted_at IS NULL
	`, sourceID)
	if err != nil {
		return nil, logDBErr(ctx, "load_full_hierarchy", sourceID, fmt.Errorf("load full hierarchy: %w", err))
	}
	defer rows.Close()

	h := &HierarchyMap{
		CityToProvince:    make(map[int64]int64),
		DistrictToCity:    make(map[int64]int64),
		SubDistrictToDist: make(map[int64]int64),
	}

	for rows.Next() {
		var pID, cID, dID, sID int64
		if err := rows.Scan(&pID, &cID, &dID, &sID); err != nil {
			return nil, logDBErr(ctx, "load_full_hierarchy_scan", sourceID, fmt.Errorf("scan hierarchy: %w", err))
		}

		h.CityToProvince[cID] = pID
		h.DistrictToCity[dID] = cID
		h.SubDistrictToDist[sID] = dID
	}

	if err := rows.Err(); err != nil {
		return nil, logDBErr(ctx, "load_full_hierarchy_rows", sourceID, fmt.Errorf("rows hierarchy: %w", err))
	}

	return h, nil
}

func (r *LocationRepository) LoadCityProvinceMapping(ctx context.Context, sourceID int64) (map[int64]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT city_id, province_id
		FROM location_hierarchy
		WHERE location_source_id = ? AND deleted_at IS NULL
	`, sourceID)
	if err != nil {
		return nil, logDBErr(ctx, "load_city_province_mapping", sourceID, fmt.Errorf("load city province mapping: %w", err))
	}
	defer rows.Close()

	mapping := make(map[int64]int64)
	for rows.Next() {
		var cityID, provinceID int64
		if err := rows.Scan(&cityID, &provinceID); err != nil {
			return nil, logDBErr(ctx, "load_city_province_mapping_scan", sourceID, fmt.Errorf("scan city province mapping: %w", err))
		}
		if _, exists := mapping[cityID]; !exists {
			mapping[cityID] = provinceID
		}
	}
	if err := rows.Err(); err != nil {
		return nil, logDBErr(ctx, "load_city_province_mapping_rows", sourceID, fmt.Errorf("rows city province mapping: %w", err))
	}
	return mapping, nil
}
