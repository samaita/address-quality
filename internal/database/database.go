package database

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

func New(dbPath string, maxOpenConns int) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Printf("database initialized: %s", dbPath)
	return &Repository{db: db}, nil
}

type AddressRecord struct {
	ID              string
	AddressID       string
	RawInput        string
	NormalizedAddr  string
	Confidence      float64
	PostalCode      string
	SubDistrict     string
	District        string
	City            string
	Province        string
	LocationVersion string
	OutputJSON      string
	CreatedAt       time.Time
}

func (r *Repository) InsertAddressRequest(ctx context.Context, rec *AddressRecord) error {
	query := `INSERT INTO address_requests (
		id, address_id, raw_input, normalized_address,
		confidence, postal_code, sub_district, district, city, province,
		location_version, output_json, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		rec.ID, rec.AddressID, rec.RawInput, rec.NormalizedAddr,
		rec.Confidence, rec.PostalCode, rec.SubDistrict, rec.District, rec.City, rec.Province,
		rec.LocationVersion, rec.OutputJSON, rec.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
