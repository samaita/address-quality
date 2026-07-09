package database

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	DB.SetMaxOpenConns(1)

	if err = DB.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if err = migrate(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Printf("database initialized: %s", dbPath)
}

func migrate() error {
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
	_, err := DB.Exec(query)
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

func InsertRecord(ctx context.Context, r *AddressRecord) error {
	query := `INSERT INTO address_requests (
		id, address_id, raw_input, normalized_address,
		confidence, postal_code, sub_district, district, city, province,
		location_version, output_json, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := DB.ExecContext(ctx, query,
		r.ID, r.AddressID, r.RawInput, r.NormalizedAddr,
		r.Confidence, r.PostalCode, r.SubDistrict, r.District, r.City, r.Province,
		r.LocationVersion, r.OutputJSON, r.CreatedAt.Format(time.RFC3339),
	)
	return err
}
