package database

import (
	"context"
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
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
