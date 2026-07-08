# Address Quality API — Indonesia

## 1. Overview

A lightweight Go HTTP server that accepts free-text Indonesian address input, sanitizes it, stores the request in SQLite, and returns a structured JSON response with quality metadata. Built for extensibility — geocoding, validation, and enrichment can be layered on top.

**Status**: v0.1.0 · MVP  
**License**: MIT

---

## 2. Quick Start

```bash
git clone <repo-url> address-quality
cd address-quality

cp .env.example .env
# edit .env as needed

make run
# or with hot reload:
# make air
```

```bash
curl -X POST http://localhost:8080/v1/address \
  -H "Content-Type: application/json" \
  -d '{"address":"Jl. Merdeka No.1, Jakarta Pusat 10110"}'
```

---

## 3. Tech Stack

| Component       | Technology                          | Version |
|----------------|--------------------------------------|---------|
| Language       | Go                                   | 1.26.5  |
| Web Framework  | Echo                                 | v4      |
| Config         | Viper (reads .env)                   | latest  |
| Sanitization   | Bluemonday (UGCPolicy)               | latest  |
| Database       | SQLite (via modernc.org/sqlite)      | latest  |
| Rate Limiting  | samaita/go-http-ratelimit            | latest  |
| Hot Reload     | Air (air-verse/air)                  | latest  |
| UUID           | google/uuid (UUIDv7)                 | latest  |

---

## 4. Configuration

All configuration is via `.env` file (not committed — see `.env.example`). Restart required after changes.

| Key           | Default | Description                      |
|---------------|---------|----------------------------------|
| `PORT`        | `8080`  | HTTP listen port                 |
| `API_KEY`     | `""`    | Reserved for future auth         |
| `RATE_LIMIT`  | `100`   | Max requests per window          |
| `RATE_WINDOW` | `60`    | Rate limit window in seconds     |

---

## 5. API Reference

### `POST /v1/address`

#### Request

```json
{
  "address": "Jl. Merdeka No.1, Jakarta Pusat 10110"
}
```

| Field     | Type   | Required | Description                  |
|-----------|--------|----------|------------------------------|
| `address` | string | yes      | Free-text Indonesian address input |

#### Response (200 OK)

```json
{
  "timestamp": "2026-07-08T12:00:00Z",
  "request_id": "01952e8e-1b40-7f47-8000-000000000001",
  "quality": {
    "address_id": "01952e8e-1b40-7f47-8000-000000000002",
    "confidence": 0.0,
    "location": {
      "postal_code": "",
      "sub_district": "",
      "district": "",
      "city": "",
      "province": ""
    },
    "normalized_input": "Jl. Merdeka No.1, Jakarta Pusat 10110",
    "output": "Jl. Merdeka No.1, Jakarta Pusat 10110",
    "location_version": "",
    "raw_input": "Jl. Merdeka No.1, Jakarta Pusat 10110"
  }
}
```

#### Error Responses

| Status | Body                                                              | Description             |
|--------|-------------------------------------------------------------------|-------------------------|
| 400    | `{"timestamp":"...","request_id":"...","error":"address is required"}` | Missing address field   |
| 429    | `{"error":"rate limit exceeded","retry_after":60}`                | Too many requests       |
| 500    | `{"timestamp":"...","request_id":"...","error":"failed to store record"}` | Database error      |

#### Rate Limit Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 42
X-RateLimit-Reset: 1718000000
```

---

## 6. Data Flow

```
Client → POST /v1/address
          │
          ▼
    Echo Router
          │
          ▼
    Rate Limiter (samaita/go-http-ratelimit)
          │
          ▼
    Handler
      ├─ Validate address non-empty
      ├─ Parse JSON → model.AddressRequest
      ├─ Generate UUIDv7 (request_id) + UUIDv7 (address_id) + timestamp
      ├─ Sanitize with bluemonday UGCPolicy
      ├─ Build quality object
      │   ├─ address_id = generated address UUID
      │   ├─ normalized_input = sanitized text
      │   ├─ output = sanitized text (placeholder)
      │   ├─ location = empty (placeholder)
      │   └─ confidence = 0.0
      ├─ Persist to SQLite (address_requests table)
      └─ Return model.AddressResponse (JSON)
```

---

## 7. Database

### Schema

```sql
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
```

### Notes

- SQLite file: `address.db` (git-ignored)
- `id` = request UUID, `address_id` = address entity UUID (separate from request ID)
- Single-writer mode (`MaxOpenConns = 1`) for SQLite safety
- Auto-migrated on startup via `CREATE TABLE IF NOT EXISTS`
- All location fields are empty in MVP — reserved for geocoding integration
- `created_at` stored as ISO8601 UTC string

---

## 8. Rate Limiting

- **Library**: `github.com/samaita/go-http-ratelimit`
- **Strategy**: Token bucket (in-memory, per IP)
- **Default**: 100 requests per 60-second window
- **Configurable** via `RATE_LIMIT` and `RATE_WINDOW` in `.env`
- **Headers**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- **Error**: HTTP 429 with JSON body

Future: Redis-backed distributed rate limiting for multi-instance deployments.

---

## 9. Architecture Decisions

| Decision | Rationale |
|----------|-----------|
| Viper for config | Standard Go config library, supports .env, env vars, defaults |
| Global config var | Simple startup-time binding; restart to reload (no hot-reload needed) |
| Pure Go SQLite (modernc) | Zero CGO dependency, easy cross-compilation |
| Rate limiter as Echo middleware | Clean separation, reusable, swappable |
| bluemonday UGCPolicy | Safe default: allows basic formatting, strips XSS |
| UUIDv7 | Time-ordered UUIDs, good for DB indexing |
| Air for dev | File watching + auto-restart, minimal config |

---

## 10. Future Features (Open for Contribution)

- [ ] **API authentication** (API key via `API_KEY` env var, header validation)
- [ ] **Health check endpoint** (`GET /health`)
- [ ] **Graceful shutdown** (signal handling)
- [ ] **Dockerfile + docker-compose**
- [ ] **CI/CD** (GitHub Actions: lint, test, build)
- [ ] **Testing suite** (unit + integration tests)
- [ ] **Structured logging** (zerolog or zap)
- [ ] **Geocoding validation** (Indonesian postal code format, known city/district/sub-district lookup, reverse geocoding)
