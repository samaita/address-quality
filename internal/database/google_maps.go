// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package database

import (
	"context"
	"database/sql"
	"time"
)

type GoogleMapsCacheRecord struct {
	RequestHash string
	Request     string
	Response    string
}

func (r *Repository) HasGoogleMapsTable(ctx context.Context) (bool, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name='google_maps_cache'`).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, logDBErr(ctx, "has_google_maps_table", "", err)
}

// GetGoogleMapsCache returns the cached response for a request hash. found is
// false when no active (non-deleted) row matches.
func (r *Repository) GetGoogleMapsCache(ctx context.Context, requestHash string) (*GoogleMapsCacheRecord, bool, error) {
	var rec GoogleMapsCacheRecord
	err := r.db.QueryRowContext(ctx, `
		SELECT request_hash, request, response
		FROM google_maps_cache
		WHERE request_hash = ? AND deleted_at IS NULL
	`, requestHash).Scan(&rec.RequestHash, &rec.Request, &rec.Response)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, logDBErr(ctx, "get_google_maps_cache", requestHash, err)
	}
	return &rec, true, nil
}

// UpsertGoogleMapsCache stores or refreshes a cached response for a request
// hash, clearing any soft-delete so the row is active again.
func (r *Repository) UpsertGoogleMapsCache(ctx context.Context, requestHash, request, response string, now time.Time) error {
	ts := now.UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO google_maps_cache (request_hash, request, response, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, NULL)
		ON CONFLICT(request_hash) DO UPDATE SET
			request = excluded.request,
			response = excluded.response,
			updated_at = excluded.updated_at,
			deleted_at = NULL
	`, requestHash, request, response, ts, ts)
	return logDBErr(ctx, "upsert_google_maps_cache", map[string]any{"request_hash": requestHash}, err)
}
