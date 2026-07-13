# Wilayah Administrasi Indonesia

Kode dan Data Wilayah Administrasi & Pulau Indonesia sesuai Kepmendagri
No 300.2.2-2430 Tahun 2025 (original source: [cahyadsn/wilayah](https://github.com/cahyadsn/wilayah.git)).

## File

- **`wilayah.sql`** — SQL dump (~92k lines) containing all administrative regions
  of Indonesia (provinces, cities/regencies, districts, sub-districts, and villages).

## Schema

```sql
CREATE TABLE wilayah (
    kode varchar(13) NOT NULL,   -- Hierarchical region code
    nama varchar(100) NOT NULL,  -- Region name
    PRIMARY KEY (kode)
);
```

## Code Hierarchy

The `kode` field encodes the administrative tree using dot-separated levels:

| Level | Length   | Example      | Description                |
|-------|----------|--------------|----------------------------|
| 1     | 2 chars  | `11`         | Province                   |
| 2     | 5 chars  | `11.01`      | City / Regency             |
| 3     | 8 chars  | `11.01.01`   | District (Kecamatan)       |
| 4     | 13 chars | `11.01.01.2001` | Village (Kelurahan/Desa) |

Example: `11.01.01.2001` → `Keude Bakongan`
(Aceh → Aceh Selatan → Bakongan → Keude Bakongan)

## Usage in Address Quality

This data powers the **v1 Administrative Tree Parser** (see project README §8).
The parser extracts administrative tokens from free-text addresses and validates
them against this reference tree to produce a confidence score.

## Loading

```bash
mysql -u <user> -p <database> < db/source/wilayah.sql
```

## Attribution

- Author: cahya dsn (cahyadsn@gmail.com)
- License: MIT
- Source: https://github.com/cahyadsn/wilayah.git
