// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"address-quality/internal/logger"
	mw "address-quality/internal/middleware"
)

func logDBErr(ctx context.Context, op string, input any, err error) error {
	if err != nil {
		logger.Error().
			Err(err).
			Str("request_id", mw.GetRequestID(ctx)).
			Str("op", op).
			Interface("input", input).
			Msg("database query error")
	}
	return err
}

type Repository struct {
	db *sql.DB
}

func New(dbPath string, maxOpenConns int) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, logDBErr(context.Background(), "open", dbPath, err)
	}

	db.SetMaxOpenConns(maxOpenConns)

	if err = db.Ping(); err != nil {
		return nil, logDBErr(context.Background(), "ping", dbPath, err)
	}

	logger.Info().Str("db_path", dbPath).Msg("database initialized")
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
	return logDBErr(ctx, "insert_address_request", map[string]any{"id": rec.ID, "raw_input": rec.RawInput}, err)
}

func (r *Repository) Ping(ctx context.Context) error {
	return logDBErr(ctx, "ping", "", r.db.PingContext(ctx))
}

func (r *Repository) HasAddressTables(ctx context.Context) (bool, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name='address_requests'`).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, logDBErr(ctx, "has_address_tables", "", err)
}

func (r *Repository) ExecSchema(ctx context.Context, sqlContent string) error {
	for _, stmt := range strings.Split(sqlContent, ";") {
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
