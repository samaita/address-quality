# Benchmark Prio: City Over Other (City-Name Collision Fix)

> **Goal:** When an address contains a token that is both a city name and a subdistrict/district name, resolve it as the **city** — unless other `place_name` evidence already pins the location. When the city is chosen and both Kota and Kabupaten forms exist, prefer **Kota** unless only Kabupaten exists.

**Architecture:** A new `location_city_priority` table (per source: normalized name + `city_type`) populated by the seeder, loaded once into memory via `sync.Once`. `city_type` (KOTA/KABUPATEN) is derived **structurally from the kode** — a city kode is `XX.YY` where the second component starting with `7` (`71`–`79`) is a Kota, anything else is a Kabupaten (verified: 514/514 level-3 rows, zero mismatches). No `LIKE` on names. All matching keys use the existing `normalizer.Normalize`. Suppression happens during entity resolution, gated on **other place-name evidence**.

**Tech Stack:** Go (service layer), SQLite (`location.db`), existing normalizer (`normalizer.Normalize` used for all keys), k6 benchmark harness.

**Current state (2026-08-20 benchmark, 106 records):** 49.1% exact matches (52/106). Challenge #2 = 25 records (23.6%).

**Root cause (verified against `db/location.db` + live API):**

- `bandung` = 2 cities (Kab. 32.04, Kota 32.73) + 2 districts + **15 subdistricts** (e.g. `35.04.17.2008` = Bandung subdist. in Kab. Tulungagung, pc 66274).
- `JL. GATOT SUBROTO NO.86 BANDUNG` → winner `Bandung/Bandung/Kab. Tulungagung/Jawa Timur 66274` (0.43, `match_district, match_subdistrict`) — the enriched subdistrict chain outscores the bare city candidate (0.37).
- **234** normalized city names also exist at level 4/5; **21** have both Kota and Kabupaten city forms.
- Verified: `Jl. Kenangan, Depok, Yogyakarta` → "yogyakarta" resolves as a city candidate (Kota Yogyakarta 0.37) but the Depok subdistrict chain (0.47) wins. `Jl. Aceh, Bandung` → "aceh" is a **road name**, resolves to nothing (sits in `ambiguous`).

---

## Design summary (user directives)

1. **New table is fine** — we cannot know which city names overlap district/subdistrict without computing it; the seeder computes and stores it.
2. **`city_type` from the kode, not the name.** A city kode is `XX.YY`; the second component `71`–`79` = Kota, otherwise Kabupaten (verified zero mismatches across all 514 cities). The seeder derives `city_type` from the kode with a substring check — no `LIKE 'Kota %'`, no dependence on the name column. All matching keys continue to use the **existing `normalizer.Normalize`** — no new normalization logic anywhere.
3. **Other-evidence detection = `place_name` evidence.** A token that resolves in the DB is evidence (existing logic). If any **other** place_name token resolves to a location entity, the city priority for the ambiguous token is **ignored**. A **resolved postal code** also counts as evidence (user directive). **Road names** resolve to nothing today (verified) — they never count, never block priority.
4. **Kota unless no Kab:** `city_type` = `KOTA` if any city row with that name is `Kota X`, else `KABUPATEN`. Used as a tie-break when the city is chosen with no other evidence.

---

## Task 1: `.gitignore` — un-ignore plan markdown files

**Objective:** `plan/*.md` must be committed for review; remove the ignore line.

**Step 1: Remove `plan/*.md` from `.gitignore`** (done — PR #1 branch)

**Step 2: Verify**

```bash
git check-ignore plan/benchmark_prio_city_over_other_level_plan.md   # should print nothing (exit 1)
```

---

## Task 2: Database — new `location_city_priority` table

**Objective:** Store, per source, city names duplicated as subdistrict/district names, plus the preferred city form type.

**Files:**
- Modify: `db/location.sql` (add table + index)
- Modify: `internal/database/location.go` (add `FindAllCityPriority` returning name + type)

**Step 1: Add table to `db/location.sql`**

```sql
-- location_city_priority: normalized city names that ALSO exist as
-- subdistrict/district names in the same source, with the preferred
-- city form type ("KOTA" or "KABUPATEN") to use when the token is
-- resolved as a city with no other disambiguating evidence.
-- city_type is derived structurally from the kode (second component
-- 71-79 = KOTA, else KABUPATEN), never from the name column.
CREATE TABLE IF NOT EXISTS location_city_priority (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    location_source_id    INTEGER NOT NULL REFERENCES location_sources(id),
    lowercase_normalized  TEXT NOT NULL,
    city_type             TEXT NOT NULL DEFAULT 'KOTA',   -- 'KOTA' | 'KABUPATEN'
    created_at            TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at            TEXT,
    deleted_at            TEXT,
    UNIQUE(location_source_id, lowercase_normalized)
);
```

**Step 2: Add repository method** in `internal/database/location.go`:

```go
type CityPriorityRow struct {
    LowercaseNormalized string
    CityType            string // "KOTA" | "KABUPATEN"
}

func (r *LocationRepository) FindAllCityPriority(ctx context.Context, sourceID int64) ([]CityPriorityRow, error) {
    rows, err := r.db.QueryContext(ctx, `
        SELECT lowercase_normalized, city_type FROM location_city_priority
        WHERE location_source_id = ? AND deleted_at IS NULL
    `, sourceID)
    if err != nil {
        return nil, logDBErr(ctx, "find_all_city_priority", sourceID, fmt.Errorf("find all city priority: %w", err))
    }
    defer rows.Close()
    var out []CityPriorityRow
    for rows.Next() {
        var r CityPriorityRow
        if err := rows.Scan(&r.LowercaseNormalized, &r.CityType); err != nil {
            return nil, logDBErr(ctx, "find_all_city_priority_scan", sourceID, fmt.Errorf("scan city priority: %w", err))
        }
        out = append(out, r)
    }
    if err := rows.Err(); err != nil {
        return nil, logDBErr(ctx, "find_all_city_priority_rows", sourceID, fmt.Errorf("rows city priority: %w", err))
    }
    return out, nil
}
```

**Step 3: Extend the `LocationRepository` interface** in `internal/service/service.go`:

```go
FindAllCityPriority(ctx context.Context, sourceID int64) ([]database.CityPriorityRow, error)
```

**Step 4: Verify build**

```bash
go build ./...
```

---

## Task 3: Seeder — populate `location_city_priority` (city_type from kode)

**Objective:** After seeding + hierarchy rebuild, insert every city-level name that also appears at level 4 or 5, with `city_type` derived **structurally from the kode** (second component `71`–`79` = KOTA, else KABUPATEN). No `LIKE` on names.

**Files:**
- Modify: `cmd/seeder/main.go`
- Modify: `internal/database/location.go` (add `RebuildCityPriority`)

**Step 1: Add repository method**

```go
func (r *LocationRepository) RebuildCityPriority(ctx context.Context, sourceID int64) error {
    if _, err := r.db.ExecContext(ctx, `
        DELETE FROM location_city_priority WHERE location_source_id = ?
    `, sourceID); err != nil {
        return logDBErr(ctx, "rebuild_city_priority_delete", sourceID, fmt.Errorf("delete city priority: %w", err))
    }
    // city_type from kode: a city kode is XX.YY; second component 71-79 = Kota,
    // anything else = Kabupaten. No name LIKE needed (verified 514/514, zero mismatches).
    _, err := r.db.ExecContext(ctx, `
        INSERT OR IGNORE INTO location_city_priority (location_source_id, lowercase_normalized, city_type)
        SELECT c3.location_source_id, c3.lowercase_normalized,
               CASE WHEN substr(c3.kode, 4, 1) = '7' THEN 'KOTA' ELSE 'KABUPATEN' END
        FROM location_codes c3
        WHERE c3.level_id = 3 AND c3.deleted_at IS NULL
          AND EXISTS (
            SELECT 1 FROM location_codes low
            WHERE low.location_source_id = c3.location_source_id
              AND low.lowercase_normalized = c3.lowercase_normalized
              AND low.level_id IN (4, 5) AND low.deleted_at IS NULL
          )
        GROUP BY c3.location_source_id, c3.lowercase_normalized
    `, sourceID)
    if err != nil {
        return logDBErr(ctx, "rebuild_city_priority_insert", sourceID, fmt.Errorf("insert city priority: %w", err))
    }
    return nil
}
```

**Note:** `substr(c3.kode, 4, 1)` reads the first digit of the second kode component (kode `32.73` → char 4 = `7`) — structurally Kota, with zero dependence on the `name` column. When a name has multiple city rows (e.g. `bandung` = `32.04` + `32.73`), the `CASE` per-row yields both `KABUPATEN` and `KOTA`, and `INSERT OR IGNORE` keeps the first — so **group by name and prefer KOTA** by ordering the aggregation. Safer variant that guarantees KOTA-wins:

```sql
SELECT c3.location_source_id, c3.lowercase_normalized,
       MAX(CASE WHEN substr(c3.kode, 4, 1) = '7' THEN 'KOTA' ELSE 'KABUPATEN' END) AS city_type
FROM location_codes c3
WHERE c3.level_id = 3 AND c3.deleted_at IS NULL
  AND EXISTS (...)
GROUP BY c3.location_source_id, c3.lowercase_normalized
```

`MAX('KOTA','KABUPATEN')` = `'KOTA'` (K > K) — the Kota-wins rule falls out of the data without any `LIKE`.

**Step 2: Call it from `cmd/seeder/main.go`** after `RebuildLocationHierarchy`:

```go
logger.Info().Msg("rebuilding city priority lookup...")
if err := repo.RebuildCityPriority(ctx, sourceID); err != nil {
    logger.Fatal().Err(err).Msg("rebuild city priority")
}
```

**Step 3: Re-seed and verify**

```bash
make build-seed
go run ./cmd/seeder --truncate
# then the normal seed path (tables must exist): go run ./cmd/seeder
```

```bash
sqlite3 db/location.db "SELECT COUNT(*) FROM location_city_priority;"               # expect 234
sqlite3 db/location.db "SELECT lowercase_normalized, city_type FROM location_city_priority WHERE lowercase_normalized IN ('bandung','depok','surabaya','karanganyar','kupang');"
# bandung -> KOTA, depok -> KOTA, surabaya -> KOTA, karanganyar -> KABUPATEN, kupang -> KOTA
```

---

## Task 4: Service — load priority set once into memory

**Objective:** Load `location_city_priority` into `map[string]string` (normalized name → city type) once per source, following the `sync.Once` cache pattern. No new normalization — the keys are already normalized by the seeder.

**Files:**
- Modify: `internal/service/service.go` (fields + interface)
- Modify: `internal/service/validate_helper.go` (loader + ensure func)
- Modify: `internal/service/resolve.go` (wire into `ensureEntitiesCachesLoaded`)

**Step 1: Add fields to `Service`**

```go
cityPrioritySet  map[string]string // normalized name -> "KOTA" | "KABUPATEN"
cityPriorityOnce sync.Once
cityPriorityErr  error
```

**Step 2: Add loader** (in `validate_helper.go`, mirroring `loadCities`):

```go
func (svc *Service) loadCityPriority(ctx context.Context, sourceID int64) {
    rows, err := svc.locationRepo.FindAllCityPriority(ctx, sourceID)
    if err != nil {
        svc.cityPriorityErr = err
        return
    }
    set := make(map[string]string, len(rows))
    for _, r := range rows {
        set[r.LowercaseNormalized] = r.CityType
    }
    svc.cityPrioritySet = set
}

func ensureCityPriorityLoaded(svc *Service, ctx context.Context, sourceID int64) error {
    svc.cityPriorityOnce.Do(func() { svc.loadCityPriority(ctx, sourceID) })
    return svc.cityPriorityErr
}
```

**Step 3: Wire into `ensureEntitiesCachesLoaded`** in `internal/service/resolve.go` (before phrase-dict build):

```go
if err := ensureCityPriorityLoaded(svc, ctx, sourceID); err != nil {
    return err
}
```

**Step 4: Verify build + loader unit test**

```bash
go build ./...
go test ./internal/service/ -run TestLoadCityPriority -v
```

---

## Task 5: Core fix — city wins, gated on other place-name evidence

**Objective:** During `ResolveEvidence`, for a priority token, suppress SUBDISTRICT/DISTRICT candidates unless another place-name (or resolved postal-code) evidence exists.

**Files:**
- Modify: `internal/service/resolve.go` (`ResolveEvidence`)
- Modify: `internal/service/validate_helper.go` (evidence-type helper)

**Step 1: Define "other evidence" check**

```go
func hasOtherLocationEvidence(resolved []model.ResolvedEvidence, priorityValue string) bool {
    for _, re := range resolved {
        if re.Evidence.Value == priorityValue {
            continue // the ambiguous token itself
        }
        // place_name evidence that resolved to a real entity pins the location
        if re.Evidence.Type == model.EvidencePlaceName && len(re.Candidates) > 0 {
            return true
        }
        // resolved postal code is also evidence (user directive)
        if re.Evidence.Type == model.EvidencePostalCode && len(re.Candidates) > 0 {
            return true
        }
        // road names never resolve today -> never count as evidence
    }
    return false
}
```

**Step 2: Apply suppression in `ResolveEvidence`** (after the resolution loop):

```go
for _, ev := range evidence {
    if ev.Type != model.EvidencePlaceName {
        continue
    }
    cityType, isPriority := svc.cityPrioritySet[ev.Value]
    if !isPriority {
        continue
    }
    if hasOtherLocationEvidence(resolved, ev.Value) {
        continue // other evidence pins the location — keep existing behavior
    }
    // only the ambiguous token is evidence -> suppress lower levels, keep CITY
    for i := range resolved {
        if resolved[i].Evidence.Value != ev.Value {
            continue
        }
        filtered := resolved[i].Candidates[:0]
        for _, c := range resolved[i].Candidates {
            if c.Level == "CITY" {
                filtered = append(filtered, c)
            }
        }
        resolved[i].Candidates = filtered
        // prefer the city_type city when ranking (KOTA over KABUPATEN)
        sort.SliceStable(resolved[i].Candidates, func(a, b int) bool {
            aType, bType := "KABUPATEN", "KABUPATEN"
            if strings.HasPrefix(resolved[i].Candidates[a].Name, "Kota ") { aType = "KOTA" }
            if strings.HasPrefix(resolved[i].Candidates[b].Name, "Kota ") { bType = "KOTA" }
            if aType != bType { return aType == cityType }
            return false
        })
    }
}
```

Note: at this point `Candidates[].Name` is the **original** name (e.g. `Kota Bandung`, `Kabupaten Bandung`) — the sort's `HasPrefix("Kota ")` reads the original name, which is correct here (this is runtime entity data, not a normalized key).

**Step 3: Explicit `kecamatan`/`kelurahan`/`desa` context**

The normalizer strips admin prefixes, so "kecamatan bandung" == "bandung" post-normalization. Detect the admin prefix **before** normalization in `ValidateAddressV1` (regex on the sanitized string: `(kecamatan|kec\.?|kelurahan|kel\.?|desa)\s+(NAME)`), collect those tokens into `explicitSubdistrictTokens`, and skip suppression for them.

**Step 4: Unit tests** (new `internal/service/priority_test.go`)

```go
func TestCityPrioritySuppressesSubdistrict(t *testing.T) {
    // "Bandung" alone; priority {"bandung":"KOTA"}
    // assert candidates contain CITY only; winner Kota Bandung
}

func TestCityPriorityIgnoredWithOtherPlaceName(t *testing.T) {
    // "Jl. Kenangan, Depok, Yogyakarta" — yogyakarta resolves as place_name
    // assert depok keeps subdistrict/district candidates (no suppression)
}

func TestCityPriorityIgnoredWithPostalCode(t *testing.T) {
    // "Depok 16431" — postal code resolves -> no suppression
    // assert resolves to Pancoran Mas, Kota Depok (unchanged)
}

func TestCityPriorityIgnoredWithProvince(t *testing.T) {
    // "Depok, Jawa Barat" — province resolves as place_name -> no suppression
    // assert resolves to Kota Depok (unchanged, must not regress)
}

func TestCityPriorityRoadNameNotEvidence(t *testing.T) {
    // "Jl. Aceh, Bandung" — aceh is road_name, resolves to nothing
    // priority applies -> bandung resolves as city (Kota Bandung)
}

func TestCityPriorityRespectsExplicitKecamatan(t *testing.T) {
    // "kecamatan bandung" — explicit context -> SUBDISTRICT entities kept
}

func TestCityPriorityKabupatenOnly(t *testing.T) {
    // karanganyar — only Kabupaten form -> city_type KABUPATEN -> Kab. Karanganyar
}
```

**Step 5: Verify**

```bash
go build ./... && go test ./internal/service/ -v
```

---

## Task 6: Benchmark — measure the improvement

**Objective:** Re-run the benchmark against the fixed build and compare challenge #2 counts.

**Step 1: Build & run the server**

```bash
go build -o bin/server ./cmd/server
# start server (existing workflow)
```

**Step 2: Run benchmark**

```bash
make benchmark
make benchmark-page
```

**Step 3: Verify expectations**

- New `tests/api/benchmark/YYYY-MM-DD_benchmark_v1_000N.json` written.
- Challenge #2 (`city`) count drops from **25**.
- `JL. GATOT SUBROTO NO.86 BANDUNG` → Kota Bandung.
- `Jl. Kenangan, Depok, Yogyakarta` → NOT Kota Depok; resolves toward Yogyakarta (priority ignored).
- `Depok, Jawa Barat` / `Depok 16431` → unchanged.
- Exact match % rises from **49.1%**.
- Check `page/benchmark.html` challenge section.

**Step 4: Regression check**

```bash
go test ./...
make test-api-smoke
```

---

## Task 7: Night dev-run schedule (cheaper off-peak)

**Objective:** All heavy development and validation runs happen at night when compute/API costs are lower.

**Context:** The plan's execution (re-seed, build, benchmark, k6 load test) is compute-heavy and hits the API repeatedly. Running these at night avoids peak pricing.

**Step 1: Define the night window**

- Suggested: **00:00–06:00 WIB** (UTC+7) — quiet hours.
- Confirm the actual cheap-rate window with the user (cloud provider billing / API tier) before hard-coding.

**Step 2: Execution order within the window**

1. Re-seed `location.db` (`bin/seeder --truncate` + seed + `RebuildCityPriority`) — heavy DB write.
2. `go build ./...` + full unit test suite.
3. Start server, run `make benchmark` + `make benchmark-page`.
4. Run k6 load test (`make test-api-load`) — the noisiest, most API-heavy step.
5. Collect results, update the benchmark page, report delta vs 49.1% baseline.

**Step 3: Automation option**

If the user wants it hands-off, schedule the whole sequence as a Hermes cron job in the night window (e.g. `0 2 * * *` WIB). The job prompt must be self-contained: re-seed, build, run benchmark, save results, report.

**Step 4: Cost note**

The night window also applies to any future benchmark/load-test cycles, not just this plan — add it to the dev routine.

---

## Acceptance Criteria

- [ ] `location_city_priority` table exists, 234 rows, `city_type` derived **from the kode** (`KOTA` for bandung/depok/surabaya/kupang — kodes `32.73`/`32.76`/`35.78`/`53.71`; `KABUPATEN` for karanganyar — kode `33.13`).
- [ ] All keys use the existing `normalizer.Normalize`; no new normalization logic.
- [ ] `Bandung` alone → **Kota Bandung**.
- [ ] `JL. GATOT SUBROTO NO.86 BANDUNG` → Kota Bandung (challenge #2 example fixed).
- [ ] `Jl. Kenangan, Depok, Yogyakarta` → priority ignored (Yogyakarta place-name evidence wins).
- [ ] `Depok, Jawa Barat` / `Depok 16431` → unchanged.
- [ ] `Jl. Aceh, Bandung` → Kota Bandung (road name is not evidence).
- [ ] `kecamatan bandung` → subdistrict path preserved.
- [ ] `karanganyar` → Kabupaten (only Kab form).
- [ ] Benchmark challenge #2 < 25; exact-match % > 49.1%.
- [ ] Dev/benchmark runs executed in the night window per Task 7.
- [ ] All Go tests + k6 smoke pass.

## Risks / Tradeoffs / Open Questions

1. **Road names are non-evidence (verified):** `Jl. Aceh, Bandung` leaves "aceh" unresolved, so priority applies. Correct per directive.
2. **Postal code counts as evidence** (user directive): a *resolved* postal code blocks priority; an *unresolved* (typo'd) one does not.
3. **KOTA is a tie-break, not an override:** priority only kicks in with no other evidence; `city_type` picks Kota among equal city candidates. Never overrides province/postal evidence.
4. **Multi-token names (25 of 234, e.g. `tebing tinggi`)** never match as single words today — inert in the priority set; no regression risk, no benefit yet.
5. **Seeding dependency:** the new table is populated by the seeder — deployed DBs need a re-seed (`bin/seeder --truncate` + seed) before the service picks it up. Include this in the night run.
6. **Night window details** (exact hours, automation vs manual) — confirm with the user in Task 7.

## Verification commands (quick reference)

```bash
go build ./... && go test ./...
make benchmark && make benchmark-page
git check-ignore plan/benchmark_prio_city_over_other_level_plan.md   # expect nothing
sqlite3 db/location.db "SELECT COUNT(*) FROM location_city_priority;"                       # 234
sqlite3 db/location.db "SELECT lowercase_normalized, city_type FROM location_city_priority WHERE lowercase_normalized='bandung';"  # KOTA
```
