# ABOUT.md

> Canonical knowledge base for the **Address Quality** repository.
> Source of truth: repository files and git history at `github.com/samaita/address-quality`.
> Classification legend: **FACT** (directly supported by repo files), **INFERENCE** (derived logically from implementation), **UNKNOWN** (insufficient evidence).
> All file references are relative to the repository root.

---

# Executive Summary

**What this project is**

- Address Quality is an HTTP API that validates and normalizes Indonesian (Bahasa Indonesia) free-text addresses against a reference administrative hierarchy, and returns structured metadata including a confidence score, quality status, resolved location components, and explainable evidence.
- FACT — README.md:3-11, internal/service/SPECS.md:1-6.
- The project consists of a Go backend (`cmd/server`, `internal/`) and a data seeding tool (`cmd/seeder`). The marketing/docs/playground frontend was removed from this repo and is now hosted at samaita.com.

**What problem it solves**

- It attempts to solve the problem of poor-quality Indonesian address data entering downstream systems (geocoders, logistics, KYC, CRM, data pipelines) by validating administrative hierarchy, resolving ambiguity, and explaining why a result is reliable.
- FACT — README.md:15-23, 62-92, 186-192.
- The intended positioning is "address intelligence layer", not a replacement for geocoders or logistics providers. README.md:66-68, 188.

**Current maturity**

- Active development ("Public Alpha"). The project is usable but backward compatibility is not guaranteed before the first stable release.
- FACT — README.md:9-11 ("Public Alpha").
- No release tags exist in git history. FACT — `git tag -l` returns empty (verified 2026-08-04).
- 130 commits by a single author across ~28 days (2026-07-08 → 2026-08-04). FACT — git history.
- The API is served behind API-key auth, per-IP per-Key rate limiting, and has both unit tests and k6 API tests. FACT — internal/router/router.go:52-57, internal/*_test.go, tests/api/*.

**Current scope**

- Indonesian administrative hierarchy: province, city/regency, district, subdistrict (kelurahan/desa), and postal code.
- FACT — internal/model/model.go:37-73; db/location.sql:1-16.
- A single validation endpoint (`POST /v1/validate`) plus health check and Swagger UI.
- FACT — internal/router/router.go:49-57.

**Current limitations**

- Road-level validation is not implemented; road-name evidence is extracted but not resolved (`resolveRoadNameEntity` always returns `nil`). FACT — internal/service/resolve.go:74-76; README.md:218.
- Geocoding / OSM / Google Maps integration is not implemented (roadmap items). FACT — README.md:219-220.
- No batch validation API. FACT — README.md:221.
- No official SDKs. FACT — README.md:222.
- English-only API documentation; input is assumed Indonesian. FACT — README.md.

Confidence: **High** (executive-level claims are directly supported by README, source, and git history).

---

# Project Vision

Inferred strictly from repository evidence:

- The project exists to reduce the cost of poor-quality Indonesian address data reaching downstream systems. README.md:21-23 states that once bad address data enters geocoders, logistics, KYC, or customer databases, correcting it becomes expensive.
- The stated goal is to become the address-intelligence layer that validates addresses *before* they reach downstream systems. README.md:62-68, 188-192.
- The design philosophy emphasizes explainability: every result should be traceable to the input evidence. README.md:156; internal/service/SPECS.md:628-631.
- The intended audience is systems where address quality affects business operations: logistics, e-commerce, fintech, insurance, KYC, CRM, data cleaning. README.md:7.
- Commercial intent is explicit: dual licensing under BUSL-1.1 with a separate Commercial License required for production/commercial use. LICENSE:1-21; COMMERCIAL_LICENSE.md:1-16.
- The frontend was deliberately kept lightweight ("vibe coded") so engineering effort stays focused on the validation engine; it has since been moved out of this repo to the main samaita.com site.

**What it intentionally does NOT solve** (per repository evidence):

- It does not geocode (no coordinate output). README.md:66-68, 90.
- It does not replace logistics platforms. README.md:66.
- It does not perform road-level validation yet. README.md:218; internal/service/resolve.go:74-76.

**Why this project exists (root cause):**

- Indonesian addresses are difficult because administrative abbreviations, aliases, landmarks, RT/RW, typos, and ambiguous names (e.g., "Bogor" could be Kota Bogor or Kabupaten Bogor) break deterministic matching. README.md:27-58.

UNKNOWN:
- The specific business motivation for choosing the Indonesian market first, and the origin story of the project, are not documented in the repository.

Confidence: **High** for stated vision (README), **Medium** for inferred intent.

---

# Repository Structure

| Path | Responsibility | Evidence |
|---|---|---|
| `cmd/server/main.go` | Backend entrypoint: loads config, opens both SQLite databases, wires sanitizer/service/handler, starts Echo server. | cmd/server/main.go:28-52 |
| `cmd/seeder/main.go` | Data seeder CLI: parses `db/source/wilayah.sql` + `wilayah_kodepos.sql`, creates schema, batch-inserts into `location.db`, rebuilds hierarchy. | cmd/seeder/main.go:22-195 |
| `internal/config/config.go` | Environment/config loading via Viper (.env + environment variables) and defaults. | internal/config/config.go:31-79 |
| `internal/database/database.go` | `address.db` repository: `address_requests` insert + ping. | internal/database/database.go:29-81 |
| `internal/database/location.go` | `location.db` repository: source lookup, level queries, hierarchy load, postal-code lookup, schema/drop/truncate/normalize operations used by seeder. | internal/database/location.go:19-582 |
| `internal/handler/handler.go` | HTTP handlers: health check, address validation. | internal/handler/handler.go:18-77 |
| `internal/logger/logger.go` | zerolog setup, Echo logger adapter, request logging middleware. | internal/logger/logger.go:20-113 |
| `internal/middleware/apikey.go` | X-API-Key authentication middleware (no-op when key is empty). | internal/middleware/apikey.go:12-26 |
| `internal/middleware/ratelimit.go` | Per-IP rate limiting wrapping `github.com/samaita/go-http-ratelimit`. | internal/middleware/ratelimit.go:16-40 |
| `internal/middleware/requestid.go` | UUID v7 request ID injection. | internal/middleware/requestid.go:17-31 |
| `internal/model/model.go` | All domain types: request/response, candidate, quality status, components, evidence, reasons, conflicts. | internal/model/model.go:11-251 |
| `internal/normalizer/normalizer.go` | Text normalization: lowercase, strip administrative prefixes, strip punctuation, collapse whitespace, preserve postal codes. | internal/normalizer/normalizer.go:11-57 |
| `internal/router/router.go` | Echo router setup: middleware chain and routes. | internal/router/router.go:22-59 |
| `internal/sanitizer/sanitizer.go` | HTML sanitization via bluemonday UGCPolicy. | internal/sanitizer/sanitizer.go:11-25 |
| `internal/service/` | Validation engine: `service.go`, `validate_v1.go`, `evidence.go`, `resolve.go`, `builder.go`, `dedup.go`, `enrich.go`, `evaluate.go`, `validate_helper.go`, `constant.go`, plus `SPECS.md` spec and tests. | internal/service/*.go |
| `docs/` | Swagger artifacts (`docs.go`, `swagger.json`, `swagger.yaml`) generated by swag; `ISSUES.md` is a gitignored architecture review. | docs/docs.go, docs/ISSUES.md:1-4 |
| `db/` | SQLite databases (`address.db`, `location.db`, both gitignored), schemas (`address.sql`, `location.sql`), and upstream data dumps (`source/wilayah.sql`, `source/wilayah_kodepos.sql`). | db/*.sql, db/source/README.md |
| `deploy/` | Production deployment artifacts: compose file, env examples, nginx example. | deploy/* |
| `tests/api/` | k6 smoke/load tests, Node benchmark script, test-case CSVs, run orchestration, result artifacts. | tests/api/run-k6.sh, tests/api/*.js |
| `.github/workflows/ci.yml` | CI: go vet/test/build, build & push the API image to GHCR. | .github/workflows/ci.yml:18-82 |
| Root scripts/config | `Makefile` (build/test/seed/benchmark targets), `Dockerfile`, `docker-compose.yml`, `deploy.sh`, `rollback.sh`, `.air.toml` (air live-reload), license files. | Makefile, Dockerfile, deploy.sh |

Confidence: **High**.

---

# Technology Stack

**Language / Runtime**

- Go (module `address-quality`, go 1.26.5). FACT — go.mod:1-3.
- Why: the whole backend is Go; build flags `CGO_ENABLED=0` for a static binary. Dockerfile:10.

**Framework**

- Echo v4 (`github.com/labstack/echo/v4 v4.15.4`) for HTTP routing and middleware. FACT — go.mod:7; internal/router/router.go:10.
- Why: direct evidence is limited; the codebase uses Echo's middleware and context APIs throughout. INFERENCE — no alternative-router rationale is documented.

**Libraries**

- `modernc.org/sqlite v1.53.0` — pure-Go SQLite driver (no CGO). FACT — go.mod:11; internal/database/database.go:11,34.
- `github.com/spf13/viper v1.21.0` — config loading from `.env` + environment. FACT — go.mod:10; internal/config/config.go:31-54.
- `github.com/rs/zerolog v1.35.1` — structured logging. FACT — go.mod:36; internal/logger/logger.go:20-38.
- `github.com/microcosm-cc/bluemonday v1.0.27` — HTML sanitization. FACT — go.mod:8; internal/sanitizer/sanitizer.go:6-17.
- `github.com/google/uuid v1.6.0` — UUID v7 for request IDs / candidate IDs / address IDs. FACT — go.mod:6; internal/middleware/requestid.go:9,21; internal/service/validate_v1.go:27.
- `github.com/samaita/go-http-ratelimit` — in-house rate limiter library. FACT — go.mod:9; internal/middleware/ratelimit.go:13.
- `swaggo` toolchain + `echo-swagger` — Swagger/OpenAPI generation and UI. FACT — go.mod:44-47; Makefile:44-45; internal/router/router.go:12,50.

**Database**

- SQLite, two separate files:
  - `db/address.db` — request log (`address_requests`). FACT — db/address.sql:1-15.
  - `db/location.db` — read-only administrative hierarchy (`location_sources`, `location_codes`, `location_alias`, `location_hierarchy`, `location_levels`). FACT — db/location.sql:1-106.
- Why SQLite: no justification is documented; the deployment mounts a plain file volume and DBs are pre-seeded on a dev machine. INFERENCE from deploy.sh and README.md:253-293.

**Infrastructure / Deployment**

- Container image: `ghcr.io/samaita/address-quality` (backend). FACT — README.md:248-250; .github/workflows/ci.yml:47-51.
- Runtime: VPS with **Podman** (rootless) + a compose provider (podman-compose ≥ 1.0.4 or docker-compose). FACT — README.md:247,267; deploy.sh:37.
- nginx: host-level reverse proxy example (`deploy/nginx.example.conf`). FACT — deploy/nginx.example.conf:6-25.
- Why Podman instead of Docker: not documented; only Podman is referenced. UNKNOWN rationale.

**Frontend**

- No frontend lives in this repo anymore. A React/TypeScript marketing + docs + playground app (Vite, Tailwind CSS v4, Cloudflare Kumo UI) was built here, then removed; the frontend is now hosted on the main samaita.com site.

**Observability**

- Structured JSON logs to stdout via zerolog; request logging middleware records request_id, method, uri, remote_ip, status, bytes, latency. FACT — internal/logger/logger.go:79-112.
- Database query errors logged with request_id and input. FACT — internal/database/database.go:17-27; git commit e2867fd.

**Testing**

- Go unit tests (5 files), k6 smoke/load API tests, a Node.js accuracy benchmark. FACT — internal/normalizer/normalizer_test.go, internal/service/*_test.go, tests/api/*.js.

**CI/CD**

- GitHub Actions: `backend-checks`, `build-push` (builds and pushes the API image to GHCR; never deploys). FACT — .github/workflows/ci.yml:18-82; README.md:247.

**AI**

- No AI/ML model is used in the runtime engine; matching is deterministic phrase/hash lookup against precomputed normalized names. INFERENCE — internal/service/validate_helper.go:276-315; no model artifacts exist.

Confidence: **High**.

---

# Architecture

## Layered architecture

```text
Client (curl / SDK-less integrations)
   │  POST /v1/validate   GET /health   GET /swagger/*
   ▼
nginx (host reverse proxy, deploy/nginx.example.conf)          [prod only]
   │  /address-quality/ → 127.0.0.1:7300
   ▼
Echo Server (cmd/server/main.go → internal/router/router.go)
   │
   ├─ logger.EchoMiddleware        (request logging)
   ├─ middleware.Recover           (panic recovery)
   ├─ middleware.CORS              (configurable origins)
   ├─ middleware.BodyLimit         (max body size)
   ├─ middleware.APIKeyAuth        (X-API-Key)
   ├─ middleware.RateLimiter       (per-IP, per-Key)
   └─ middleware.RequestID         (UUID v7)
   │
   ▼
Handler (internal/handler/handler.go)
   │
   ▼
Service — ValidateAddressV1 (internal/service/validate_v1.go)
   │  8-stage pipeline (SPECS.md)
   ▼
LocationRepository (internal/database/location.go) + AddressRepository (internal/database/database.go)
   │
   ▼
SQLite: db/location.db (hierarchy) / db/address.db (request log)
```

## In-process validation pipeline (the engine)

```text
Raw Address
   → sanitize (bluemonday)
   → normalize (normalizer.Normalize)
   → Evidence Extraction (ExtractEvidence)
   → Entity Resolution (ResolveEvidence + matchPhrases)
   → Candidate Discovery (DiscoverCandidates / pathBuilder)
   → Candidate Deduplication (DeduplicateCandidates)
   → Candidate Enrichment (EnrichCandidates — fills uppers from hierarchy)
   → Conclusion Building (BuildConclusions — attach matched evidence)
   → Candidate Evaluation (EvaluateCandidate — hierarchy/coverage/conflicts/score/status)
   → Ranking & Winner Selection
   → Postal-code resolution fallback
   → Response formatting + persistence (address_requests)
```

FACT — internal/service/validate_v1.go:21-199; internal/service/SPECS.md:32-60, 604-650.

## Runtime data caching

- The Service lazily loads entire administrative levels into in-memory maps guarded by `sync.Once` (provinces, cities, districts, subdistricts, hierarchy, phrase dictionary). FACT — internal/service/service.go:61-96; internal/service/validate_helper.go:196-274.
- Because caches load with `sync.Once` keyed to the first `sourceID`, all requests share one in-memory snapshot after first load. INFERENCE — internal/service/validate_helper.go:196-219.

Confidence: **High**.

---

# Request Lifecycle

Detailed lifecycle of `POST /v1/validate`:

1. **TLS / reverse proxy** (production): host nginx terminates TLS and proxies `/address-quality/` to `127.0.0.1:7300`, forwarding `X-Forwarded-*` and websocket upgrade headers. FACT — deploy/nginx.example.conf:6-25.
2. **Echo server** starts the request; read/write timeouts from config. FACT — internal/router/router.go:27-28.
3. **Logging middleware** — begins timing, writes one access log line per request at the end. FACT — internal/logger/logger.go:79-112.
4. **Recover middleware** — catches panics. FACT — internal/router/router.go:31.
5. **CORS middleware** — applies configured allowed origins/methods/headers. FACT — internal/router/router.go:32-47.
6. **BodyLimit** — rejects bodies over `MAX_BODY_SIZE` (default "1M"). FACT — internal/router/router.go:53.
7. **APIKeyAuth** — if `API_KEY` is configured, compares `X-API-Key` header; 401 JSON on mismatch. If `API_KEY` is empty, passes through. FACT — internal/middleware/apikey.go:12-26.
8. **RateLimiter** — per-IP (parsed from `RemoteAddr`) with rule `* /v1/*`; on exceed returns 429 with `Retry-After` and JSON body. FACT — internal/middleware/ratelimit.go:16-40.
9. **RequestID** — injects UUID v7 into context; used for logging and responses. FACT — internal/middleware/requestid.go:17-31.
10. **Handler bind** — `c.Bind(&req)`; malformed JSON → 400 `ErrorResponse`. FACT — internal/handler/handler.go:60-66.
11. **Input validation** — `req.Validate(maxAddressLength)`: address required, max length check; failure → 400. FACT — internal/model/model.go:27-35; internal/service/validate_v1.go:22-24.
12. **Address ID** — UUID v7 generated for the record. FACT — internal/service/validate_v1.go:27.
13. **Sanitization** — bluemonday `UGCPolicy` sanitize. FACT — internal/service/validate_v1.go:28; internal/sanitizer/sanitizer.go:15-25.
14. **Normalization** — lowercase, strip administrative prefixes, strip punctuation, collapse spaces, preserve postal codes. FACT — internal/normalizer/normalizer.go:33-57.
15. **Source resolution** — look up `location_sources` by `source_code` (request value or default); 500 on failure. FACT — internal/service/validate_v1.go:32-40.
16. **Cache warm** — load provinces/cities/districts/subdistricts/hierarchy (lazy, once). FACT — internal/service/validate_v1.go:42-45; internal/service/resolve.go:78-94.
17. **Evidence extraction** — find 5-digit postal code, then per-word road-name/place-name evidence. FACT — internal/service/evidence.go:15-56.
18. **Entity resolution** — match phrases against the phrase dictionary; postal codes resolved against subdistrict cache; road names resolve to nothing. FACT — internal/service/resolve.go:12-76.
19. **Candidate discovery** — build candidates from province/city/district/subdistrict entities (top-down + any-level strategies). FACT — internal/service/builder.go:21-41.
20. **Deduplication** — merge candidates with identical location keys. FACT — internal/service/dedup.go:10-29.
21. **Enrichment** — fill upper levels from hierarchy cache; attach postal code from subdistrict. FACT — internal/service/enrich.go:11-65.
22. **Conclusion building** — assign matched evidence to each candidate. FACT — internal/service/builder.go:117-167.
23. **Evaluation** — hierarchy validation, completeness, evidence coverage, conflict detection, confidence scoring, status assessment, reasons. FACT — internal/service/evaluate.go:13-32.
24. **Ranking** — sort by confidence desc, then level count desc, then fewer conflicts. FACT — internal/service/validate_v1.go:64-74.
25. **Status override** — if top two candidates are close (< 0.1 apart) and top is VALID, status becomes AMBIGUOUS. FACT — internal/service/validate_v1.go:76-86.
26. **Postal-code fallback** — if the winner is empty, infer full location from postal code DB; if winner lacks postal/subdistrict, fill postal code. FACT — internal/service/validate_helper.go:29-67.
27. **Serialization** — build `ResponseData` with matched/missing/conflicts/ambiguous, resolution candidates, metadata. FACT — internal/service/validate_v1.go:151-174.
28. **Persistence** — insert one row into `address_requests`; DB error → 500. FACT — internal/service/validate_v1.go:187-190; internal/handler/handler.go:68-74.
29. **Response** — `AddressResponse` JSON with timestamp and request_id. FACT — internal/service/validate_v1.go:192-196.

**Authentication** — API key only on `/v1/*` group; `/health` and `/swagger/*` are public. FACT — internal/router/router.go:49-57.

**Rate limiting** — applies to the `api` group (all `/v1/*`); health check is *not* rate-limited in current code. FACT — internal/router/router.go:52-56 (differs from the gitignored ISSUES.md which reviewed an older revision where `/health` sat behind the limiter — ISSUES.md A5/C1; current code contradicts that stale review).

Confidence: **High**.

---

# Domain Model

All concepts are defined in `internal/model/model.go` unless noted.

## Administrative hierarchy

- **Province** — top level (level_id 2). Model: `Province{ID, Name, NormalizedName}`. model.go:189-193.
- **City / Regency** — level_id 3, includes `PostalCode`. `City{ID, Name, NormalizedName, PostalCode}`. model.go:195-201. Includes both Kota and Kabupaten types; the normalizer strips both `kota` and `kabupaten` prefixes. normalizer.go:11-27.
- **District** (kecamatan) — level_id 4. `District{ID, Name, NormalizedName}`. model.go:203-207.
- **Subdistrict** (kelurahan/desa) — level_id 5, carries the postal code. `SubDistrict{ID, Name, NormalizedName, PostalCode}`. model.go:208-213. (git commit ff45967 renamed "Village" → "SubDistrict".)
- **Postal Code** — 5-digit; attached to subdistricts in data, but also modeled as its own entity. model.go:215-219; db/location.sql:48.
- **Road** — modeled (`Road{ID, Name}`) but never populated at runtime. model.go:221-224; internal/service/resolve.go:74-76.

Relationship encoding: the Kemendagri `kode` is dot-separated (province.city.district.subdistrict), e.g. `11.01.01.2001`. FACT — db/source/README.md:36-50. A precomputed `location_hierarchy` table maps each subdistrict to its province/city/district uppers. FACT — db/location.sql:85-106.

## Candidate

- A candidate is one possible interpretation of the address — a partially populated `AdminLocation` plus evidence and discovery metadata. `AdminCandidate{UUID, OriginLevel, Location, Evidence, DiscoveryStrategies}`. model.go:245-251.
- `AdminLocation{Province, City, District, SubDistrict, PostalCode, Road, Conflicts}`. model.go:226-234.
- `OriginLevel` ∈ PROVINCE/CITY/DISTRICT/SUBDISTRICT — the level the candidate was built from. model.go:236-243.

## Evidence

- Raw facts extracted from the normalized text, not yet assigned to hierarchy. model.go:158-161; internal/service/evidence.go:15-56.
- Types: `road_name`, `place_name`, `postal_code`. model.go:150-156.
- Resolved evidence: evidence + list of possible entities. `ResolvedEvidence{Evidence, Candidates}`. model.go:170-173.
- Matched evidence: evidence + single resolved entity attached to a candidate. `MatchedEvidence{Evidence, Resolved}`. model.go:184-187.

## Discovery strategies

- Enum: `top_down`, `any_level`, `alias`, `postal`. model.go:175-182. Only `top_down` and `any_level` are passed at runtime; `alias` and `postal` are declared but not used in the current pipeline. FACT — model.go:177-182; internal/service/validate_v1.go:53.

## Quality status

- `VALID`, `INCOMPLETE`, `AMBIGUOUS`, `CONFLICT`, `UNKNOWN`. model.go:55-63; produced by `assessQuality` (evaluate.go:361-382) with an AMBIGUOUS override at the top-two-candidates step (validate_v1.go:76-86).

## Components

- `province`, `city`, `district`, `sub_district`, `postal_code`. model.go:65-73. Used for matched/missing arrays.

## Conflict types

- `hierarchy_conflict`, `postal_code_mismatch`, `orphan_city`, `multiple_city`, `duplicate_level`. FACT — internal/service/evaluate.go:79-207.

## Reason constants

- `exact_match`, `match_province`, `match_city`, `match_district`, `match_subdistrict`, `match_postal_code_exact`, `match_postal_code_prefix4`, `match_postal_code_prefix3`, `postal_code_lookup`, `postal_code_inferred`. model.go:80-93.

## Relationships

- Province → City → District → Subdistrict (strict parent-child, validated against `HierarchyMap`). evaluate.go:79-103.
- Postal code belongs to a subdistrict; conflicts flagged when a candidate's postal code does not match its subdistrict. evaluate.go:134-141.
- Evidence → Entity → Candidate: multiple entities can resolve from one evidence; one entity can appear in multiple candidates. builder.go:43-63.

Confidence: **High**.

---

# API Overview

Public endpoints (FACT — internal/router/router.go:49-57):

## GET /health

- **Purpose**: health check; returns 503 if the address database ping fails.
- **Authentication**: none.
- **Rate limiting**: none (not in the `/v1/*` group).
- **Request**: none.
- **Response 200**: `{"status":"ok","database":"ok"}`; **Response 503**: `{"status":"error"}`. FACT — internal/handler/handler.go:32-38.
- **Evidence**: internal/handler/handler.go:32-38; internal/router/router.go:49.
- Note: only `address.db` is pinged (`svc.Ping` → `repo.Ping`), not `location.db`. FACT — internal/service/service.go:102-104; internal/handler/handler.go:33.

## GET /swagger/*

- **Purpose**: Swagger UI served by `echo-swagger`. FACT — internal/router/router.go:50.
- **Authentication**: none. FACT — internal/router/router.go:50 (outside protected group).
- **Evidence**: internal/router/router.go:50; docs/swagger.yaml.

## POST /v1/validate

- **Purpose**: validate and resolve an Indonesian address. FACT — internal/handler/handler.go:49-59.
- **Authentication**: `X-API-Key` header, required when `API_KEY` is configured; 401 otherwise. FACT — internal/middleware/apikey.go:12-26; docs/swagger.yaml:225-229.
- **Rate limiting**: per-IP, `RATE_LIMIT` per `RATE_WINDOW` seconds (rule `* /v1/*`). FACT — internal/middleware/ratelimit.go:26-28.
- **Request body**:
  ```json
  { "address": "JL. Supratman No.72, Citarum, 40191", "source_code": "kemendagri" }
  ```
  `address` required; `source_code` optional (defaults to config `LOCATION_SOURCE_CODE`). FACT — internal/model/model.go:22-25; internal/service/validate_v1.go:32-35; docs/swagger.yaml:3-13.
- **Validation**:
  - Body must be valid JSON (400 otherwise). handler.go:63-66.
  - `address` required and ≤ `MAX_ADDRESS_LENGTH` chars (400 otherwise). model.go:27-35.
  - Body size ≤ `MAX_BODY_SIZE` (middleware). router.go:53.
- **Response 200** — `AddressResponse`:
  ```json
  {
    "timestamp": "…", "request_id": "…",
    "data": {
      "address_id": "…", "status": "VALID",
      "confidence": 0.35,
      "raw_input": "…", "normalized_input": "…", "formatted_address": "…",
      "location": {"postal_code":"…","sub_district":"…","district":"…","city":"…","province":"…"},
      "assessment": {"matched":[], "missing":[], "conflicts":[], "ambiguous":[]},
      "resolution": {"strategy":[], "candidate_count":1, "candidates":[{uuid,score,location,reasons}]},
      "metadata": {"location_source":"kemendagri","location_version":"2025"}
    }
  }
  ```
  FACT — internal/model/model.go:126-148; docs/swagger.yaml:144-173.
- **Response 400** — `ErrorResponse` (invalid body / validation error). handler.go:64-72.
- **Response 401** — `{"error":"invalid or missing API key"}`. apikey.go:19-21.
- **Response 429** — `{"error":"rate limit exceeded","retry_after":N}` with `Retry-After` header. ratelimit.go:31-36.
- **Response 500** — `ErrorResponse` ("failed to store record") on source lookup or DB insert failure. handler.go:68-74.
- **Evidence**: internal/handler/handler.go:60-77; internal/router/router.go:57; docs/swagger.yaml:194-224.

## Note on documentation drift

- Swagger descriptions and the `@description` in cmd/server/main.go:6 say "Thai address validation and resolution API" — a copy-paste artifact; the product validates Indonesian addresses. FACT — cmd/server/main.go:4-8; docs/swagger.yaml:177,198. (The `@host localhost:8080` in main.go:7 also differs from the configured default port 7300.)

Confidence: **High**.

---

# Data Model

## Database 1: `db/address.db`

- Single table `address_requests` (FACT — db/address.sql:1-15):
  - `id TEXT PRIMARY KEY` — the request_id (UUID v7 string).
  - `address_id TEXT NOT NULL` — distinct per-request UUID v7.
  - `raw_input TEXT NOT NULL`, `normalized_address TEXT`, `confidence REAL`, `postal_code`, `sub_district`, `district`, `city`, `province`, `location_version`, `output_json TEXT`, `created_at TEXT`.
- No indexes beyond the PK. FACT — db/address.sql.
- Written by `Repository.InsertAddressRequest`. FACT — internal/database/database.go:65-77.
- Current row count on disk: 1,594 (verified via sqlite3, 2026-08-04). FACT — db/address.db (on-disk state; file is gitignored).

## Database 2: `db/location.db`

Tables (FACT — db/location.sql:1-106):

- `location_levels(id, code UNIQUE, name)` — seeded rows: 1 country, 2 province, 3 city, 4 district, 5 subdistrict. db/location.sql:5-16.
- `location_sources(id, code, version, name, code_date, desc, created_at, updated_at, deleted_at)` with `UNIQUE(code, version)` — registry of upstream datasets. db/location.sql:28-39.
- `location_codes(id, location_source_id FK, kode, name, lowercase_normalized, level_id FK, postal_code, created_at, updated_at, deleted_at)` with `UNIQUE(location_source_id, kode)` and index `location_codes_normalized_idx(lowercase_normalized)`. db/location.sql:52-66.
- `location_alias(id, location_id FK, alias, …)` with `UNIQUE(location_id, alias)` — declared but never populated at runtime by the current code/seeder. FACT — db/location.sql:75-83 (seeder only inserts sources/codes/hierarchy; alias table exists but no writes).
- `location_hierarchy(id, location_source_id FK, province_id FK, city_id FK, district_id FK, subdistrict_id FK, …)` with `UNIQUE(location_source_id, subdistrict_id)` — precomputed upper chain per subdistrict. db/location.sql:95-106.

## Indexes

- `location_codes_normalized_idx` on `lowercase_normalized` (db/location.sql:66) — used for phrase lookup.
- Unique constraints: `(code, version)` sources; `(location_source_id, kode)` codes; `(location_id, alias)` aliases; `(location_source_id, subdistrict_id)` hierarchy.
- `address_requests` has no secondary indexes. FACT — db/address.sql.

## Constraints

- Foreign keys via `REFERENCES` (not enforced unless PRAGMA foreign_keys is enabled — the code does not enable it). INFERENCE — db/location.sql uses `REFERENCES`; no `PRAGMA foreign_keys` found in `internal/database/location.go`.
- Soft delete via `deleted_at` on sources/codes/hierarchy; all queries filter `deleted_at IS NULL`. FACT — db/location.sql:27,51,74,94; internal/database/location.go:58,79,115,454,490,526.

## Migration strategy

- No migration framework. Schema is applied by the seeder (`cmd/seeder --init` reads `db/location.sql`), and `address.db` has no migration code at all. FACT — cmd/seeder/main.go:94-106; db/address.sql.
- `cmd/seeder` flags: `--init` (create schema), `--drop` (drop all), `--truncate` (clear rows), `--normalize` (rebuild `lowercase_normalized`). cmd/seeder/main.go:28-32,54-72.

## Repository pattern

- Two repositories: `Repository` (address.db) and `LocationRepository` (location.db), constructed in main.go. FACT — internal/database/database.go:29-47; internal/database/location.go:19-37.
- The Service depends on interfaces `AddressRepository` and `LocationRepository` (service.go:18-33), enabling test doubles (used in `enrich_test.go`). FACT — internal/service/service.go:18-33; internal/service/enrich_test.go.

## Seeder data flow

1. Parse `db/source/wilayah.sql` (~91,790 lines) → `(kode, name)`; level inferred by counting dots (`level_id = dots + 2`). FACT — cmd/seeder/main.go:197-226.
2. Parse `db/source/wilayah_kodepos.sql` (~84,058 lines) → `kode → kodepos` map. FACT — cmd/seeder/main.go:238-263.
3. Join postal codes to rows by `kode`. FACT — cmd/seeder/main.go:166-173.
4. Batch insert (500/batch) into `location_codes`, normalizing names on insert. FACT — cmd/seeder/main.go:175-186; internal/database/location.go:367-391.
5. Rebuild `location_hierarchy` from subdistrict codes. FACT — cmd/seeder/main.go:190-194; internal/database/location.go:191-290.

Current on-disk hierarchy counts (verified via sqlite3, 2026-08-04): 38 provinces, 514 cities, 7,285 districts, 83,762 subdistricts; one source row `(kemendagri, 2025, "Kepmendagri No 300.2.2-2138")`. FACT — db/location.db (gitignored file state).

Confidence: **High**.

---

# Validation Pipeline

Execution order, per `SPECS.md` and `internal/service/validate_v1.go`:

1. **Sanitize** — bluemonday `UGCPolicy` HTML sanitize. FACT — validate_v1.go:28.
2. **Normalize** — `normalizer.Normalize`: lowercase; drop administrative prefixes (`kabupaten`, `kab`, `kota`, `kecamatan`, `kec`, `kelurahan`, `kel`, `provinsi`, `prov`, `administrasi`, `kepulauan`); remove non-alpha punctuation; collapse whitespace; preserve 5-digit postal codes (re-appended). FACT — normalizer.go:11-57.
3. **Input validation** — required + max length. FACT — model.go:27-35.
4. **Evidence extraction** — regex `\b(\d{5})\b` for postal code; strip road prefixes `jl|jalan|gg|gang`; per remaining word, classify as `road_name` vs `place_name`; de-duplicate by lowercase word. FACT — evidence.go:13-56. (Note: `isRoadNameWord` currently returns `false` for every word, so all non-postal evidence is emitted as `place_name`. FACT — evidence.go:58-65.)
5. **Entity resolution** — for each evidence:
   - Postal code → all subdistricts with that code. resolve.go:39-54.
   - Place name → entities from phrase dictionary keyed `"sourceID:phrase"`. resolve.go:56-72.
   - Road name → `nil` (unimplemented). resolve.go:74-76.
   - Phrase matching uses longest-match over word sequences. FACT — validate_helper.go:276-315.
6. **Candidate discovery** — build one candidate per unique entity at each of PROVINCE/CITY/DISTRICT/SUBDISTRICT; postal evidence excluded from candidate roots. FACT — builder.go:21-63.
7. **Candidate deduplication** — merge by location key (province:city:district:subdistrict ids), union strategies and evidence. FACT — dedup.go:10-109.
8. **Enrichment** — bottom-up upper fill (subdistrict→district→city→province) from hierarchy; attach subdistrict postal code. FACT — enrich.go:11-65.
9. **Conclusion building** — attach matched evidence to candidates; sort by level count. FACT — builder.go:117-167.
10. **Evaluation** (per candidate):
    - Hierarchy validation: city↔province, district↔city, subdistrict↔district via `HierarchyMap`; mismatches → `hierarchy_conflict`. evaluate.go:79-110.
    - Completeness: matched/missing component lists. evaluate.go:34-69.
    - Evidence coverage: unused evidence (values not matched by the candidate). evaluate.go:115-132.
    - Conflict detection: postal/subdistrict mismatch, orphan city, multiple cities, duplicate levels. evaluate.go:134-207.
    - Confidence scoring. evaluate.go:237-288.
    - Quality status. evaluate.go:361-382.
    - Reasons. evaluate.go:384-424.
11. **Ranking & selection** — confidence desc, then filled-level count desc, then conflict count asc. validate_v1.go:64-74.
12. **Ambiguity check** — top two within 0.1 confidence and top is VALID → AMBIGUOUS. validate_v1.go:76-86.
13. **Postal-code fallback** — winner empty → infer from postal code; winner lacks postal/subdistrict → fill postal code. validate_helper.go:29-67.
14. **Formatting** — formatted address = `subdistrict, district, city, province [postal]`. validate_v1.go:229-241.

Confidence: **High**.

---

# Scoring Logic

Implemented in `internal/service/evaluate.go` (`scoreConfidence`) and `internal/service/constant.go`.

## Evidence weights (constant.go:6-21)

| Signal | Weight |
|---|---|
| Exact match (any resolved evidence) | 0.10 |
| Hierarchy valid + city present | 0.15 |
| Postal code present | 0.05 |
| Province matched | 0.15 |
| City matched | 0.12 |
| District matched | 0.08 |
| Subdistrict matched | 0.05 |
| Multi-evidence per province | 0.20 |
| Multi-evidence per city | 0.15 |
| Multi-evidence per district | 0.10 |
| Multi-evidence per subdistrict | 0.05 |
| Max multi-evidence bonus cap | 0.40 |
| Postal prefix ≥3 match | 0.07 |
| Postal prefix ≥4 match | 0.10 |
| Postal prefix 5 (exact) match | 0.20 |

FACT — constant.go:6-21; evaluate.go:237-288.

## Scoring mechanics (evaluate.go:237-288)

- Sum of applicable weights; `hierarchy_conflict` suppresses the hierarchy weight (0.15). evaluate.go:251-261.
- Postal prefix scoring compares input postal digits against the candidate's postal code and awards tiered weights for ≥3/≥4/5 leading-char matches. evaluate.go:290-321.
- `multiEvidenceBonus` counts how many evidence items resolve to the same entity and adds the per-level bonus, capped at 0.40. evaluate.go:323-359.
- Score is clamped to [0, 1] and rounded to 4 decimals. evaluate.go:284-287.

## Status derivation (assessQuality, evaluate.go:361-382)

- Any conflict → `CONFLICT`.
- No administrative component → `UNKNOWN`.
- Missing province, city, or district → `INCOMPLETE`.
- Otherwise → `VALID`.
- `AMBIGUOUS` is only set at the response level when the top two ranked candidates are within 0.1 confidence and the top is VALID. FACT — validate_v1.go:76-86.

## Penalties

- No explicit negative weights exist; penalties are implicit via conflict-driven status demotion (CONFLICT) and via suppressed hierarchy weight on hierarchy conflicts. INFERENCE — evaluate.go:251-261,361-382.

## Ranking / tie-breaking

- Sort key: confidence desc → count of non-nil location levels desc → conflict count asc. FACT — validate_v1.go:64-74.

## Confidence in tests

- evaluate_test.go asserts exact expected scores (e.g., exact-only 0.10, hierarchy-only 0.15, full pipeline 0.37, multi-evidence combinations 0.52/0.60). FACT — internal/service/evaluate_test.go:255,429,710 (see test report).

UNKNOWN:
- The rationale for choosing these specific weight values is not documented (only `SPECS.md` states confidence "is implementation-specific"). FACT — SPECS.md:472-487.

Confidence: **High** for mechanics, **Low** for why weights were chosen.

---

# Middleware

In execution order for `/v1/*` (FACT — internal/router/router.go:30-56; logger.EchoMiddleware is registered globally at router.go:30 and applies to all routes):

1. **logger.EchoMiddleware** (global) — access logging to zerolog; level Info, or Warn for 4xx, Error for 5xx; fields: request_id, method, uri, remote_ip, host, status, bytes_in, bytes_out, latency. Purpose: observability. FACT — internal/logger/logger.go:79-112.
2. **middleware.Recover** (global, from Echo) — panic recovery. FACT — router.go:31.
3. **middleware.CORS** (global) — configurable origins/methods/headers, MaxAge 3600. Purpose: allow browser clients (playground). FACT — router.go:32-47.
4. **middleware.BodyLimit** (api group) — caps request body at `MAX_BODY_SIZE`. Purpose: DoS protection. FACT — router.go:53.
5. **APIKeyAuth** (api group) — X-API-Key check when configured; 401 otherwise. Purpose: authentication. FACT — router.go:54; apikey.go:12-26.
6. **RateLimiter** (api group) — per-IP, rule `* /v1/*`, 429 + Retry-After. Purpose: abuse protection. FACT — router.go:55; ratelimit.go:16-40.
7. **RequestID** (api group) — UUID v7 into context. Purpose: traceability. FACT — router.go:56; requestid.go:17-31.

**Tradeoffs observed:**

- The rate-limiter key is parsed from `RemoteAddr` by stripping the last colon — IPv6 addresses and trusted-proxy handling are not accounted for. INFERENCE — ratelimit.go:19-25.
- `BodyLimit` parses a string like "1M" (Echo's format). FACT — router.go:53; config default "1M" (config.go:43).
- Health check intentionally bypasses auth and rate limiting by living outside the `api` group. FACT — router.go:49 vs 52-57.
- Request ID is generated even for failed API-key / rate-limit requests? No — RequestID runs *after* APIKeyAuth and RateLimiter, so unauthorized/rate-limited responses do not carry a request_id in the ErrorResponse body (handler reads it from context; those middleware return their own JSON). INFERENCE — middleware order at router.go:54-56; errorResponse at handler.go:40-47 is only used by the handler itself.

Confidence: **High**.

---

# Configuration

## Environment variables (config.Load, internal/config/config.go:14-78)

| Env var | Default | Purpose |
|---|---|---|
| `PORT` | 7300 | HTTP listen port |
| `API_KEY` | "" | Shared API key; empty disables auth |
| `RATE_LIMIT` | 100 | Max requests per window |
| `RATE_WINDOW` | 60 | Window seconds |
| `READ_TIMEOUT` | 5 | Server read timeout (s) |
| `WRITE_TIMEOUT` | 10 | Server write timeout (s) |
| `MAX_BODY_SIZE` | "1M" | Max request body |
| `MAX_ADDRESS_LENGTH` | 1000 | Max address characters |
| `ADDRESS_DB_PATH` | "db/address.db" | Path to request-log DB |
| `LOCATION_DB_PATH` | "db/location.db" | Path to hierarchy DB |
| `DB_MAX_OPEN_CONNS` | 10 | SQLite max open conns |
| `LOCATION_SOURCE_CODE` | "kemendagri" | Default data source code |
| `LOG_LEVEL` | "info" | zerolog level |
| `CORS_ALLOWED_ORIGINS` | "https://samaita.com" | Comma-separated allowed origins |

FACT — config.go:31-71; .env.example:1-12.

## Configuration loading

- Viper reads `.env` (env type) in the working directory; if missing, warns and uses defaults; `AutomaticEnv` allows env-var overrides. FACT — config.go:32-54.
- The `.env` file is gitignored (`.gitignore:1`); `.env.example` is tracked. FACT — .gitignore:1; git ls-files.
- Production uses `/etc/address-quality/.env.prod` mounted via `env_file` in compose. FACT — deploy/docker-compose.prod.yml:7-8; deploy/.env.prod.example.

## Secrets

- `API_KEY` is the only secret; stored in `.env` / `.env.prod` (chmod 0600 per README.md:279-284). FACT — README.md:279-284.
- The local `.env` (gitignored) contains a real-looking UUID API key value; `.env.prod.example` uses `CHANGE_ME`. FACT — .env:2; deploy/.env.prod.example:2.

## Runtime configuration

- All settings are load-time; no hot reload or feature-flag framework exists in the backend. FACT — config.go (no reload code).

## Build configuration

- `Makefile` targets: run, build, test, lint (`go vet`), clean, air, swagger, test-api, test-api-smoke, test-api-load, build-seed, seed, benchmark. FACT — Makefile:1-41.
- `.air.toml` configures the air live-reload dev server (pre_cmd runs swag init). FACT — .air.toml:1-18.

Confidence: **High**.

---

# Deployment

## Binary / image build

- Backend built statically: `CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server`; runtime `alpine:3.21`, non-root `appuser`, HEALTHCHECK via wget on `/health`, `EXPOSE 7300`. FACT — Dockerfile:1-28.
- Local dev compose (`docker-compose.yml`): maps 7300, mounts `address-data` volume, uses `.env`. FACT — docker-compose.yml:1-13.

## Production topology (FACT — README.md:246-312; deploy/docker-compose.prod.yml:1-22)

```text
Internet
   → (host nginx reverse proxy, deploy/nginx.example.conf)   server_name api.samaita.com
       /address-quality/ → 127.0.0.1:7300
   → Podman container:
       api       ghcr.io/samaita/address-quality:{API_IMAGE_TAG}   port 7300
```

- API container: `mem_limit 512m`, `cpus 1.0`, `security_opt no-new-privileges`, healthcheck, env_file `/etc/address-quality/.env.prod`, volume `/etc/address-quality/db:/data/db`. FACT — deploy/docker-compose.prod.yml:2-22.
- SSL: no TLS config is present in the repo (host nginx example listens on plain 80). INFERENCE — deploy/nginx.example.conf:6-12. TLS termination is presumed to be external or unconfigured.
- Rootless podman needs unprivileged port 80 (`net.ipv4.ip_unprivileged_port_start=80`). FACT — README.md:289-294.

## Host layout (FACT — README.md:249-258)

```
/etc/address-quality/
├── .env.prod                  (backend env, 0600)
├── deploy.sh / rollback.sh
├── docker-compose.prod.yml
└── db/
    ├── address.db
    └── location.db            (created/uploaded manually)
```

## Database provisioning

- DBs are never stored in images and are not managed by CI; they are seeded on a dev machine (`bin/seeder --init && bin/seeder`) and uploaded via scp, then chowned to UID 1000 (`appuser`). FACT — README.md:280-287.

## Health checks

- Container HEALTHCHECK wget `/health`; deploy.sh waits up to 60s for `healthy`, then verifies the nginx→API proxy (`/address-quality/health`) works. FACT — Dockerfile:25-26; deploy/docker-compose.prod.yml:13-18; deploy.sh:70-86.

## Release flow

- CI builds and pushes the API image to GHCR with tags `latest` (default branch), `sha-<short>` and `v*`/`v{MAJOR}.{MINOR}`. FACT — .github/workflows/ci.yml:63-71.
- **CI never deploys.** FACT — README.md:247; .github/workflows/ci.yml (no deploy job).

## Deployment flow

- Operator runs `/etc/address-quality/deploy.sh` (optionally pinning `API_IMAGE_TAG`); script pulls, recreates containers with `--force-recreate`, waits for health, verifies the nginx→API proxy, records the tag to `.release`, prunes dangling images; serializes via `flock`. FACT — deploy.sh:33-91.

## Rollback strategy

- `/etc/address-quality/deploy.sh --rollback` (or `rollback.sh`) redeploys the tag recorded in `$CONFIG_DIR/.release`. FACT — deploy.sh:54-58; rollback.sh:1-4.
- Rollback is image-tag based; there is no automated DB migration/rollback. INFERENCE — no migration tooling exists in the repo.

Confidence: **High**.

---

# Security

**Authentication**

- Single shared API key via `X-API-Key` header; when `API_KEY` is empty the middleware is a no-op (auth disabled). FACT — internal/middleware/apikey.go:12-26.
- Only `/v1/*` requires the key; `/health` and `/swagger/*` are public. FACT — internal/router/router.go:49-57.

**Authorization**

- None. There is no per-key scoping, roles, or ownership; all valid keys are equal. FACT — apikey.go.

**API keys**

- One static key per deployment via env. Rotation requires changing `.env.prod` and redeploying. INFERENCE — config.go:16,60.

**Rate limiting**

- Per-IP per-Key limiting on `/v1/*`; 429 with `Retry-After`. FACT — ratelimit.go:16-40.
- Limitation: IP is parsed from `RemoteAddr` by trimming after last colon; behind a proxy the real client IP would come from `X-Forwarded-For`, which the limiter does not read. INFERENCE — ratelimit.go:19-25. (The host nginx proxy does set `X-Forwarded-For`, deploy/nginx.example.conf:16.)

**Input validation**

- Body size cap, address required + length cap. FACT — router.go:53; model.go:27-35.
- HTML sanitization (bluemonday) applied to the address before normalization. FACT — sanitizer.go; validate_v1.go:28. Note: UGCPolicy is HTML-oriented for a plain-text field — flagged in the stale ISSUES.md S4. FACT — sanitizer.go:15-17.

**Headers**

- No security headers (CSP, HSTS, etc.) are configured in the repo. INFERENCE — no such headers found in nginx configs or Echo setup.

**Known protections**

- Non-root container user (`appuser`), `no-new-privileges:true`, resource limits (mem/cpu), ReadTimeout/WriteTimeout. FACT — Dockerfile:14-24; deploy/docker-compose.prod.yml:19-22; router.go:27-28.

**Known gaps (from repository evidence, not external review)**

- No request logging of failed auth attempts beyond access logs. INFERENCE — logger middleware logs status, not auth specifics.
- Swagger UI is exposed publicly (no auth). FACT — router.go:50.

Confidence: **High** for what exists; the security posture section for gaps is INFERENCE.

---

# Testing

**Go unit tests (5 files, all white-box / same-package)**

- `internal/normalizer/normalizer_test.go` — `TestNormalize` (19 cases) and `TestNormalizeCSV` (normalizes `tests/api/cases/example.csv`; hard dependency on that fixture path). FACT — normalizer_test.go:13,50; git commit d9d3652.
- `internal/service/builder_test.go` — candidate builders + `BuildConclusions` grouping (31 candidates grouped by level count). FACT — builder_test.go:88-288.
- `internal/service/enrich_test.go` — 8 enrichment cases (upper fill, nil hierarchy, unknown ids). FACT — enrich_test.go:44-241.
- `internal/service/evaluate_test.go` — hierarchy conflicts, completeness, conflicts, confidence weights, statuses, reasons, evidence coverage, multi-evidence bonus, postal prefix scoring (865 lines, table-driven). FACT — evaluate_test.go:21-865.
- `internal/service/phrase_test.go` — longest-match phrase resolution, source-ID isolation, multi-level phrases, determinism. FACT — phrase_test.go:97-312.

**API tests (k6)**

- `tests/api/smoke-test.js` — single VU/iteration; checks `/health`, valid address response fields, HTML sanitization, empty-address 400, malformed-body 400; threshold `checks rate==1.0`. FACT — smoke-test.js:6-92.
- `tests/api/load-test.js` — 10 default VUs, ramp 10s/hold 30s/ramp-down 10s; 5 hardcoded addresses round-robin; thresholds `http_req_failed < 0.01`, `checks == 1.0`; custom `success_duration` trend. FACT — load-test.js:7-49.
- `tests/api/run-k6.sh` — orchestrates k6 with CSV output into `tests/api/result/YYYY-MM-DD_<name>_NNNN.csv`. FACT — run-k6.sh:4-22.

**Benchmarks**

- `tests/api/benchmark-test.js` — Node script replaying `address-tagged.csv` ground truth (Source;SERP;Address Raw;Province;City;District;Subdistrict, semicolon-delimited) through the API, comparing returned province/city/district/subdistrict case-insensitively, writing JSON result files (`YYYY-MM-DD_benchmark_v1_NNNN.json`). FACT — benchmark-test.js:5-243.
- Run via `make benchmark` (`--source=kemendagri --csv=tests/api/cases/address-tagged.csv`). FACT — Makefile:47-48.
- Historical artifacts show CSV (07-20), XLSX (07-21), then JSON (07-23/24) output formats. FACT — tests/api/benchmark/*. The benchmark README documents the older CSV/XLSX format (drift). FACT — tests/api/benchmark/README.md.

**Test data**

- `tests/api/cases/example.csv` — tracked, 7 columns, 4 rows (Bandung, source "Scribd"). FACT — tests/api/cases/example.csv; .gitignore:9-10.
- `tests/api/cases/address.csv` (3 columns, 115 rows) and `address-tagged.csv` (7 columns, 124 rows) — gitignored. FACT — .gitignore:9; agent-verified row counts.

**Coverage**

- No coverage profile command exists (no `-cover`, no `coverprofile` in Makefile/CI). INFERENCE — no such flags found.
- Root `package.json` `test` script is a stub that always errors. FACT — package.json scripts.test.

**CI enforcement**

- CI runs `go vet ./...`, `go test ./...`, and build smoke for server+seeder. FACT — .github/workflows/ci.yml:30-37.

Confidence: **High**.

---

# Engineering Decisions

| Decision | Reason (inferable) | Benefits | Tradeoffs | Evidence |
|---|---|---|---|---|
| **Echo (not Gin/stdlib)** | Echo's middleware + routing fit the layered middleware design; no rationale documented | Built-in middleware, CORS, body limit, context APIs | Extra dependency; framework-specific APIs | go.mod:7; router.go:10-18 |
| **SQLite (two files, not PostgreSQL)** | No external DB service; file volume mounted on VPS; DBs seeded off-host | Zero-ops, no CGO driver, easy backup | Single-writer concurrency limits; no network DB | go.mod:11; database.go:11,34; README.md:253-264 |
| **Two databases (request log vs hierarchy)** | Separation of mutable log data from static reference data | Hierarchy DB is read-only & immutable; log DB is append-only | Two files to manage/seed | db/address.sql; db/location.sql |
| **modernc.org/sqlite (pure Go)** | Enables `CGO_ENABLED=0` static binary | Trivially deployable in Alpine | Slightly slower than CGO sqlite; memory usage | Dockerfile:10; go.mod:11 |
| **API-key authentication** | Single shared key protects the public `/v1/*` route | Simple, stateless | Shared key has no scoping; key is static | apikey.go:12-26 |
| **Per-IP per-Key rate limiting via in-house library** | `github.com/samaita/go-http-ratelimit` | Configurable rules | IP parsing naive (RemoteAddr) | ratelimit.go:16-40; go.mod:9 |
| **8-stage pipeline / SPECS.md-driven design** | SPECS.md explicitly defines single-responsibility stages; refactor commits cite it (e169ba0) | Deterministic, explainable candidates | Heavier per-request compute | SPECS.md:32-60; git log |
| **Hierarchical candidate discovery** | Address validity depends on administrative hierarchy, not string similarity | Detects internal inconsistency (conflicts) | Requires full dataset in memory | builder.go; evaluate.go:79-103; README.md:121 |
| **Weighted confidence scoring** | Weights tuned empirically (WIP scoring commits); values in constant.go | Comparable, testable score | Weight rationale undocumented | constant.go:6-21; evaluate_test.go |
| **In-memory full-dataset cache (`sync.Once`)** | Dataset small enough (38/514/7285/83762 rows) | Fast lookups, loaded once | Source-code changes require restart to reload | service.go:61-96; validate_helper.go:196-274 |
| **POST-only single validate endpoint** | Minimal MVP surface | Simple | No batch, no GET reference, no updates | router.go:57; README.md:221 |
| **Podman + GHCR, CI builds images only** | Manual operator deploys; images immutable | Reproducible deploys, no CI-to-prod path | Manual deploy step required | README.md:248; ci.yml; deploy.sh |
| **BUSL-1.1 + commercial licensing** | Monetization of validation engine | Protects IP while allowing eval use | Restricts self-host/production use | LICENSE; COMMERCIAL_LICENSE.md |
| **Swaggo-generated Swagger** | Auto-generated from handler annotations | Keeps API docs in sync with code | Docs drift visible in "Thai address" copy | main.go:4-8; Makefile:44-45 |

Confidence: **High** for facts, **Medium** for inferred reasons.

---

# Observed Project Evolution

Inferred from git history (130 commits, single author, linear history, no tags, 2026-07-08 → 2026-08-04):

**Phase 1 — Bootstrap & MVP (07-08)**
- Initial commit (`.gitignore` + 1-line README), then MVP Go API with config/database/handler/middleware/model/sanitizer; endpoint renamed `/v1/address` → `/v1/validate` on day one. FACT — git log.

**Phase 2 — API hardening (07-09)**
- Health endpoint with DB ping, default port 7300, k6 smoke/load suite + CSV wrapper, per-IP rate limiting, context propagation, DI refactor (removed globals), API-key auth, max body/address caps, structured quality output. FACT — git log.

**Phase 3 — Containerization & first CI (07-10 → 07-11)**
- Dockerfile, docker-compose, GH Actions→GHCR + Podman deploy script, nginx reverse-proxy example, UUID v7 request IDs in middleware, service layer extraction. FACT — git log.

**Phase 4 — Data layer & seeder (07-13 → 07-14)**
- `location.db` schema, `wilayah.sql` data, postal codes, `lowercase_normalized`, seeder binary with `--init/--drop/--truncate/--normalize`, hierarchy rebuild, zerolog structured logging, caches. FACT — git log (e.g., 8572f59, be25337, 26352fc).

**Phase 5 — Caching & candidate sets (07-15)**
- Map-based caches, n-gram matching, postal fallback, hierarchy filtering, admin-prefix stripping in normalizer, "Candidate Set Resolver", rename Village→SubDistrict, normalizer consolidation. FACT — git log.

**Phase 6 — Benchmark, docs, licensing (07-20 → 07-21)**
- Benchmark script + README, mermaid diagrams, BUSL-1.1 + commercial license; example.csv gained ground-truth columns. (Note: binaries were accidentally committed and later removed.) FACT — git log.

**Phase 7 — Evaluation engine overhaul (07-23 → 07-28)**
- Candidate Evaluation Engine per SPECS.md, 8-stage pipeline rewrite (builder/dedup/evidence/resolve), WIP "Scoring Issue Fix" commits, dead-code cleanup, postal prefix scoring, longest-match phrases, UUID v7 per candidate, OriginLevel, Reason constants. FACT — git log.

**Phase 8 — API documentation (07-28 → 07-29)**
- Swaggo + ErrorResponse type + swagger.json/yaml + example tags. FACT — git log (4840ca3, c9cce49).

**Phase 9 — Frontend (07-29 → 07-31, removed 08-05)**
- Five WIP commits bootstrap frontend (Vite/React/Tailwind), landing page, playground, docs page; then Public Alpha landing overhaul, playground redesign, Tailwind v4 + Kumo UI migration, AQ_API_* env rename, CORS, pagination, docs rebuild to mirror api.md with mermaid. The frontend was later removed from this repo and is now hosted at samaita.com. FACT — git log.

**Phase 10 — Production CI/CD polish (08-03 → 08-04)**
- Production CI/CD for backend (deploy.sh, rollback.sh, compose.prod), secrets-context fixes, build/push-only CI, nginx proxy fix, DB error logging, app→api rename. FACT — git log.

**Branches**

- `main` (130 commits) — sole production line, linear, no merge commits.
- `candidates` (73 commits) — experiment branch diverging at 07-23 with one unique WIP commit (3cad8cb, ±855 lines in validate_helper.go) that was never merged; represents abandoned scoring-engine rework. FACT — git history.
- No tags, no releases. FACT — `git tag -l` empty.

Confidence: **High** (based on commit messages + file changes; phase boundaries are interpretive).

---

# Known Limitations

Only limitations supported by repository evidence:

1. **Road-level validation unimplemented** — `resolveRoadNameEntity` returns `nil`; roadmap marks road-level validation as not started. FACT — internal/service/resolve.go:74-76; README.md:218.
2. **No geocoding / OSM / Google Maps integration**. FACT — README.md:219-220.
3. **No batch validation API**. FACT — README.md:221.
4. **No official SDKs**. FACT — README.md:222.
5. **No backward-compatibility guarantee before first stable release**. FACT — README.md:11.
6. **No TLS configuration in repo** (host nginx example is HTTP). INFERENCE — deploy/nginx.example.conf:6-12.
7. **Rate limiter uses `RemoteAddr` (proxy unaware)**. INFERENCE — ratelimit.go:19-25.
8. **Health check pings only `address.db`, not `location.db`**. FACT — handler.go:33; service.go:102-104.
9. **`location_alias` table exists but is never populated or queried at runtime**. FACT — db/location.sql:75-83; no writes found in code.
10. **Alias/postal discovery strategies declared but unused**; only `top_down` and `any_level` are active. FACT — model.go:177-182; validate_v1.go:53.
11. **All evidence is currently classified as `place_name`** because `isRoadNameWord` always returns false. FACT — evidence.go:58-65.
12. **Documentation drift**: swagger says "Thai" (cmd/server/main.go:6); README cites dataset "No300.2.2-2430" while seeded data/`wilayah.sql` header say "No 300.2.2-2138" (README.md:178 vs db/location.db + db/source/wilayah.sql header). FACT.
13. **Benchmark README drift**: documents CSV/XLSX output but the script writes JSON. FACT — tests/api/benchmark/README.md vs benchmark-test.js:235.
14. **Single shared API key, no scoping**. FACT — apikey.go.
15. **No coverage reporting / no coverage gate**. INFERENCE — no coverage commands in Makefile/CI.
16. **Rate limiter 429 body `retry_after` derived from window, but the window is configurable; body/header match current config at startup.** FACT — ratelimit.go:31-36 (uses `windowSec`).

Confidence: **High** (each item traceable to files).

---

# Future Directions

Inferred only from explicit markers in repository evidence:

**Roadmap (README.md:211-223)** — status is a FACT of the roadmap document, not of implementation:
- [x] Administrative hierarchy parser (implemented — FACT: builder.go, SPECS.md)
- [x] Candidate generation (implemented — FACT: builder.go)
- [x] Confidence scoring (implemented — FACT: evaluate.go)
- [x] Explainable evidence (implemented — FACT: evidence.go)
- [x] Postal code validation (implemented — FACT: evaluate.go:290-321)
- [ ] Road-level validation (placeholder)
- [ ] OpenStreetMap integration (placeholder)
- [ ] Google Maps fallback (placeholder)
- [ ] Batch validation API (placeholder)
- [ ] Official SDKs (placeholder)

**Open interfaces / stubs that anticipate work**:
- `Road` model + `resolveRoadNameEntity` stub. FACT — model.go:221-224; resolve.go:74-76.
- `location_alias` table ready for aliases. FACT — db/location.sql:75-83.
- Unused discovery strategies `alias` and `postal`. FACT — model.go:177-182.
- Unused `evaluateCompleteness` body (empty). FACT — evaluate.go:112-113.

**Comments**: No TODO/FIXME/HACK/XXX comments exist in backend source or tests. FACT — agent-verified across all source files.

**Licensing future**: "Future major versions may be released under different licensing terms." FACT — LICENSE:18-20.

Confidence: **High** for roadmap items; **Medium** for inferred intent.

---

# Canonical Vocabulary

| Term | Definition | Related concepts | Where used | Preferred wording |
|---|---|---|---|---|
| **Address Quality** | The product/API; an address intelligence layer for Indonesian addresses | — | README.md | "Address Quality API" |
| **Administrative hierarchy / administrative levels** | country(1)→province(2)→city(3)→district(4)→subdistrict(5), encoded in dot-separated `kode` | Province, City, District, SubDistrict, PostalCode | db/location.sql; normalizer.go | "administrative hierarchy" |
| **Candidate** | One possible interpretation of the input, a partially-filled `AdminLocation` with evidence | AdminCandidate, OriginLevel, DiscoveryStrategy | builder.go; dedup.go; evaluate.go | "candidate" |
| **Evidence** | A fact extracted from the normalized text (road_name/place_name/postal_code) | ResolvedEvidence, MatchedEvidence | evidence.go; model.go:150-187 | "evidence" |
| **Entity resolution** | Mapping evidence to all possible administrative entities; never picks a winner | ResolvedEvidence, phraseDict | resolve.go; SPECS.md:157-218 | "entity resolution" |
| **Discovery strategy** | How a candidate was found (top_down, any_level, alias, postal) | DiscoveryStrategy | builder.go; model.go:175-182 | "discovery strategy" |
| **Confidence** | A 0–1 score (weighted sum, clamped, 4-decimal) of candidate reliability | scoreConfidence, weights | evaluate.go:237-288; constant.go | "confidence" |
| **Quality status** | VALID / INCOMPLETE / AMBIGUOUS / CONFLICT / UNKNOWN | QualityStatus | model.go:55-63; evaluate.go:361-382 | "status" |
| **Conflict** | Internal inconsistency in a candidate (hierarchy, postal, orphan, multiple, duplicate) | Conflict types | evaluate.go:71-207 | "conflict" |
| **Component** | province/city/district/sub_district/postal_code | Matched/Missing | model.go:65-73 | "component" |
| **Reason** | Human-readable justification strings (`match_province`, `match_postal_code_prefix4`, …) | Reason constants | model.go:80-93; evaluate.go:384-424 | "reason" |
| **Normalization** | Deterministic transform: lowercase, strip administrative prefixes, strip punctuation, keep postal codes | normalizer.Normalize | normalizer.go | "normalize" |
| **Subdistrict** | Level-5 administrative unit (kelurahan/desa); renamed from "Village" in history | SubDistrict | model.go:208-213; git log ff45967 | "sub_district" |
| **Postal code / kodepos** | 5-digit code; attached to subdistricts in data | PostalCode, FindByPostalCode | db/source/README.md; database/location.go:393-440 | "postal_code" |
| **Hierarchy map** | In-memory parent lookup: city→province, district→city, subdistrict→district | HierarchyMap, location_hierarchy | database/location.go:509-555 | "hierarchy" |
| **Formatted address** | `subdistrict, district, city, province [postal]` string | formatLocation | validate_v1.go:229-241 | "formatted_address" |
| **request_id / address_id** | UUID v7 identifiers; request_id per HTTP request, address_id per validation record | middleware.RequestID | requestid.go; validate_v1.go:27 | "request_id" |

Confidence: **High**.

---

# Interesting Engineering Stories

Story candidates (title, why interesting, evidence, audience, difficulty, follow-ups). Not written — only identified.

1. **"Validating Indonesian addresses without geocoding: the 8-stage evidence pipeline"**
   - Why: how the engine avoids string-matching traps (abbreviations, aliases, ambiguous city names) using evidence → entities → candidate graphs; SPECS.md defines the philosophy.
   - Evidence: internal/service/SPECS.md:1-698; validate_v1.go:21-199.
   - Audience: engineers building matching/validation systems.
   - Difficulty: Medium. Follow-ups: comparison with geocoder-based validation; how explainability is serialized.

2. **"One address, many Bogor: the ambiguity problem and AMBIGUOUS status"**
   - Why: ties between top candidates within 0.1 confidence flip status; explains why "Bogor" alone can't be resolved.
   - Evidence: validate_v1.go:76-86; evaluate.go:361-382; README.md:38.
   - Audience: PMs/engineers. Difficulty: Low. Follow-ups: tie-breaking policy, thresholds.

3. **"Shipping the whole Indonesian administrative tree in a SQLite file: the seeder"**
   - Why: parsing ~176k lines of MySQL dumps into a normalized SQLite hierarchy with zero CGO.
   - Evidence: cmd/seeder/main.go:197-263; db/source/wilayah.sql (91,790 lines); db/location.sql.
   - Audience: data engineers. Difficulty: Medium. Follow-ups: dataset refresh workflow, checksums/verification.

4. **"SQLite as the production database for a read-heavy validation API"**
   - Why: two-file split (immutable reference vs append-only log), full-dataset in-memory cache, single-writer considerations.
   - Evidence: db/address.sql; db/location.sql; service.go:61-96; database.go.
   - Audience: backend/infra engineers. Difficulty: Medium. Follow-ups: cache invalidation, source-switch behavior, benchmarks.

5. **"A scoring model you can unit-test: confidence weights in a table"**
   - Why: weights live in one file and are asserted to 2-decimal exactness in tests; the WIP "Scoring Issue Fix" saga shows the calibration loop.
   - Evidence: constant.go:6-21; evaluate_test.go:255-765; git log 6b7997a, b6647e1.
   - Audience: ML/scoring engineers. Difficulty: Medium. Follow-ups: weight sensitivity analysis, benchmark-driven tuning.

6. **"CI that builds but never deploys: immutable images + an operator-run deploy script"**
   - Why: security-oriented split (no CI-to-prod path), `sha-` tags for reproducibility, rollback file.
   - Evidence: .github/workflows/ci.yml:70-120; deploy.sh:62-105; README.md:248.
   - Audience: platform/SRE engineers. Difficulty: Low. Follow-ups: zero-downtime rollback, secret rotation.

7. **"The candidates branch that never was: an abandoned 855-line rework"**
   - Why: a full parallel validation rewrite was abandoned on a side branch; main took a different path (SPECS.md-driven).
   - Evidence: git branch `candidates`; commit 3cad8cb; linear main history with zero merges.
   - Audience: engineers. Difficulty: Low. Follow-ups: process learnings, why approaches diverge.

Confidence: **High** for evidence cited.

---

# Questions For The Author

Questions that cannot be answered from repository evidence:

1. Why was this project started? What is the origin story and the motivation behind focusing on Indonesian addresses?
2. Why were these specific confidence weights chosen (0.10 exact, 0.15 hierarchy/province, 0.05 postal, etc.)? Was there a calibration dataset or manual tuning process?
3. Why SQLite (with two files) instead of PostgreSQL or a hosted geocoding service?
4. Why Podman (and podman-compose) instead of Docker?
5. Why was the `candidates` branch abandoned, and what was learned from it?
6. Why is the rate limiter per-IP based on `RemoteAddr` rather than `X-Forwarded-For`? Is the production deployment behind a load balancer?
7. What is the intended long-term roadmap beyond the README checklist? What is the target for the "first stable release" and backward-compatibility guarantee?
8. The README cites "Keputusan Menteri Dalam Negeri No300.2.2-2430 2025" but the seeded data and source dumps say "No 300.2.2-2138". Which is authoritative, and is a data refresh planned?
9. What data was used to evaluate accuracy (the gitignored `address.csv` / `address-tagged.csv` sets)? Where did the ground truth come from (e.g., Scribd links)?
10. Is `/health`'s lack of a `location.db` check intentional?
11. Are there plans to populate the `location_alias` table and enable the `alias`/`postal` discovery strategies?
12. What is the target capacity / performance envelope (the load tests use 10 VUs; is that representative of production traffic)?
13. What is the commercial licensing process and the pricing/positioning of the Commercial License?

---

*Document generated from repository evidence on 2026-08-04. Git revision: main @ 6106e41 (`chore: rename app to api`).*
