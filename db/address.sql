CREATE TABLE IF NOT EXISTS address_requests (
    id                 TEXT PRIMARY KEY,
    address_id         TEXT NOT NULL,
    raw_input          TEXT NOT NULL,
    normalized_address TEXT DEFAULT '',
    confidence         REAL DEFAULT 0,
    postal_code        TEXT DEFAULT '',
    sub_district       TEXT DEFAULT '',
    district           TEXT DEFAULT '',
    city               TEXT DEFAULT '',
    province           TEXT DEFAULT '',
    location_version   TEXT DEFAULT '',
    output_json        TEXT DEFAULT '',
    created_at         TEXT NOT NULL
);
