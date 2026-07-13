-- location_levels: lookup table mapping level_id to administrative hierarchy levels
--   id        - primary key, referenced by location_codes.level_id
--   code      - numeric hierarchy level (1=country, 2=province, 3=city, 4=district, 5=subdistrict)
--   name      - human-readable level name
CREATE TABLE IF NOT EXISTS location_levels (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    code INTEGER NOT NULL UNIQUE,
    name TEXT NOT NULL
);

INSERT OR IGNORE INTO location_levels (id, code, name) VALUES
    (1, 1, 'country'),
    (2, 2, 'province'),
    (3, 3, 'city'),
    (4, 4, 'district'),
    (5, 5, 'subdistrict');

-- location_sources: registry of upstream data sources (e.g. official government datasets)
--   id         - primary key, referenced by location_codes.location_source_id
--   code       - short identifier for the source dataset
--   version    - version tag of the dataset
--   name       - human-readable source name
--   code_date  - effective date of the source codes
--   desc       - description of the source dataset
--   created_at - row creation timestamp
--   updated_at - last update timestamp (NULL if never updated)
--   deleted_at - soft-delete timestamp (NULL = active)
CREATE TABLE IF NOT EXISTS location_sources (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    code       TEXT NOT NULL,
    version    TEXT NOT NULL,
    name       TEXT NOT NULL,
    code_date  TEXT,
    desc       TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT,
    deleted_at TEXT,
    UNIQUE(code, version)
);

-- location_codes: core table storing all administrative region records
--   id                 - primary key
--   location_source_id - FK to location_sources (which dataset this record came from)
--   kode               - hierarchical region code (dot-separated, e.g. "11.01.01.2001")
--   name               - official region name (e.g. "Kab. Bandung", "Kota Jakarta")
--   lowercase_normalized - lowercased name with abbreviations expanded and punctuation removed, used for full-text / fuzzy search indexing
--   level_id           - FK to location_levels (country/province/city/district/subdistrict)
--   postal_code        - 5-digit Indonesian postal code (only on subdistrict-level rows)
--   created_at         - row creation timestamp
--   updated_at         - last update timestamp (NULL if never updated)
--   deleted_at         - soft-delete timestamp (NULL = active)
CREATE TABLE IF NOT EXISTS location_codes (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    location_source_id    INTEGER NOT NULL REFERENCES location_sources(id),
    kode                  TEXT NOT NULL,
    name                  TEXT NOT NULL,
    lowercase_normalized  TEXT NOT NULL DEFAULT '',
    level_id              INTEGER NOT NULL REFERENCES location_levels(id),
    postal_code           TEXT,
    created_at            TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at            TEXT,
    deleted_at            TEXT,
    UNIQUE(location_source_id, kode)
);

CREATE INDEX IF NOT EXISTS location_codes_normalized_idx ON location_codes(lowercase_normalized);

-- location_alias: alternative names / aliases for locations (e.g. alternate spellings, abbreviations)
--   id          - primary key
--   location_id - FK to location_codes(id)
--   alias       - alternative name string
--   created_at  - row creation timestamp
--   updated_at  - last update timestamp (NULL if never updated)
--   deleted_at  - soft-delete timestamp (NULL = active)
CREATE TABLE IF NOT EXISTS location_alias (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    location_id INTEGER NOT NULL REFERENCES location_codes(id),
    alias       TEXT NOT NULL,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT,
    deleted_at  TEXT,
    UNIQUE(location_id, alias)
);
