# Wilayah Administrasi Indonesia

Kode dan Data Wilayah Administrasi & Pulau Indonesia sesuai Kepmendagri
No 300.2.2-2430 Tahun 2025 (original source: [cahyadsn/wilayah](https://github.com/cahyadsn/wilayah.git)).

## File

- **`wilayah.sql`** — SQL dump (~92k lines) containing all administrative regions
  of Indonesia (provinces, cities/regencies, districts, sub-districts, and subdistricts).
- **`wilayah_kodepos.sql`** — SQL dump (~84k lines) containing postal codes (`kodepos`)
  mapped to subdistrict/subdistrict-level region codes (kode). Both datasets share the
  same source and use `kode` as the common key.

## Schema

### wilayah (administrative regions)

```sql
CREATE TABLE wilayah (
    kode varchar(13) NOT NULL,   -- Hierarchical region code
    nama varchar(100) NOT NULL,  -- Region name
    PRIMARY KEY (kode)
);
```

### wilayah_kodepos (postal codes)

```sql
CREATE TABLE wilayah_kodepos (
    kode   varchar(13) NOT NULL,  -- Hierarchical region code (matches wilayah.kode)
    kodepos varchar(5) DEFAULT NULL,  -- 5-digit postal code
    PRIMARY KEY (kode)
);
```

## Code Hierarchy

The `kode` field encodes the administrative tree using dot-separated levels:

| Level | Length   | Example          | Description                |
|-------|----------|------------------|----------------------------|
| 1     | 2 chars  | `11`             | Province                   |
| 2     | 5 chars  | `11.01`          | City / Regency             |
| 3     | 8 chars  | `11.01.01`       | District (Kecamatan)       |
| 4     | 13 chars | `11.01.01.2001`  | SubDistrict (Kelurahan/Desa)   |

Example: `11.01.01.2001` → `Keude Bakongan`
(Aceh → Aceh Selatan → Bakongan → Keude Bakongan)

Example with postal code: `11.01.01.2001` → `Keude Bakongan` → `23773`

## Usage in Address Quality

This data powers the **v1 Administrative Tree Parser**.
The parser extracts administrative tokens from free-text addresses and validates
them against this reference tree to produce a confidence score. Postal codes from
`wilayah_kodepos` are joined via `kode` to enrich the `Location` output.

## Loading

The admin region and postal code data are pre-loaded into the SQLite `location.db`
database during setup. The import process:

1. Transform `db/source/wilayah.sql` (MySQL syntax) into SQLite-compatible INSERT
   statements and load into `location_codes` / `location_levels` / `location_sources`.
2. Transform `db/source/wilayah_kodepos.sql` (MySQL syntax) into SQLite-compatible
   UPDATE statements:

   ```sql
   UPDATE location_codes SET postal_code = '<kodepos>' WHERE kode = '<kode>';
   ```

## Attribution

- Author: cahya dsn (cahyadsn@gmail.com)
- License: MIT
- Source: https://github.com/cahyadsn/wilayah.git
