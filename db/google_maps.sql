-- google_maps_cache: lazy cache of Google Maps geocoding responses keyed by
-- a hash of the canonicalized (trimmed, single-line) request body.
--   request_hash - sha256 of the trimmed single-line request body (primary key)
--   request      - the trimmed single-line request body used to compute the hash
--   response     - raw Google Maps geocoding JSON (live, cached, or mocked)
--   created_at   - row creation timestamp
--   updated_at   - last update timestamp
--   deleted_at   - soft-delete timestamp (NULL = active)
CREATE TABLE IF NOT EXISTS google_maps_cache (
    request_hash TEXT PRIMARY KEY,
    request      TEXT NOT NULL,
    response     TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL,
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS google_maps_cache_created_at_idx ON google_maps_cache(created_at);
