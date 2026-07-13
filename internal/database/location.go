package database

import (
	"context"
	"database/sql"
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
