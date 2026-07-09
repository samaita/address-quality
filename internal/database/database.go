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

func New(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	if err = migrate(db); err != nil {
		return nil, err
	}

	log.Printf("database initialized: %s", dbPath)
	return &Repository{db: db}, nil
}

func migrate(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS address_requests (
		id                TEXT PRIMARY KEY,
		address_id        TEXT NOT NULL,
		raw_input         TEXT NOT NULL,
		normalized_address TEXT DEFAULT '',
		confidence        REAL DEFAULT 0,
		postal_code       TEXT DEFAULT '',
		sub_district      TEXT DEFAULT '',
		district          TEXT DEFAULT '',
		city              TEXT DEFAULT '',
		province          TEXT DEFAULT '',
		location_version  TEXT DEFAULT '',
		output_json       TEXT DEFAULT '',
		created_at        TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	return err
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

func (r *Repository) InsertRecord(ctx context.Context, rec *AddressRecord) error {
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
