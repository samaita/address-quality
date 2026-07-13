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

CREATE TABLE IF NOT EXISTS location_code (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    location_source_id INTEGER NOT NULL REFERENCES location_sources(id),
    kode               TEXT NOT NULL,
    name               TEXT NOT NULL,
    level_id           INTEGER NOT NULL REFERENCES location_levels(id),
    created_at         TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at         TEXT,
    deleted_at         TEXT,
    UNIQUE(location_source_id, kode)
);

CREATE TABLE IF NOT EXISTS location_alias (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    location_id INTEGER NOT NULL REFERENCES location_code(id),
    alias       TEXT NOT NULL,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT,
    deleted_at  TEXT,
    UNIQUE(location_id, alias)
);
