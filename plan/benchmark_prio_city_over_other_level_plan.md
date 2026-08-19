# Benchmark Prio: City Over Other (City-Name Collision Fix)

> **Goal:** When an address contains a token that is both a city name and a subdistrict/district name, resolve it as the **city** — unless other evidence (province, district, subdistrict, postal code) already pins the location. When the city is chosen and both Kota and Kabupaten forms exist, prefer **Kota** unless only Kabupaten exists.

**Architecture:** A new in-memory lookup (`cityPrioritySet`, `map[string]string{normalized_name: "KOTA"|"KABUPATEN"}`) loaded once at startup from a new `location_city_priority` table (generated during seeding). The priority applies **only when the token resolves to a subdistrict/district AND no other resolved location evidence exists in the address**. Explicit `kecamatan`/`kelurahan`/`desa` context always overrides the priority.

**Tech Stack:** Go (service layer), SQLite (`location.db`), existing normalizer + k6 benchmark harness.

**Current state (2026-08-20 benchmark, 106 records):** 49.1% exact matches (52/106). Challenge #2 = 25 records (23.6%).

**Root cause (verified against `db/location.db` + live API on localhost:7300):**

- `bandung` = 2 cities (Kab. 32.04, Kota 32.73) + 2 districts + **15 subdistricts** (e.g. `35.04.17.2008` = Bandung subdist. in Kab. Tulungagung, pc 66274).
- `JL. GATOT SUBROTO NO.86 BANDUNG` → winner `Bandung/Bandung/Kab. Tulungagung/Jawa Timur 66274` (0.43, reasons `match_district, match_subdistrict`). The subdistrict-origin chain gets enriched to 4 levels and outscores the bare city candidate (0.37, `match_city` only).
- **234 normalized city names also exist at level 4/5.** 209 are single-token (matchable as words); 25 are multi-token (e.g. `tebing tinggi`, `tanjung pinang` — these never match as single words today, priority is moot for them).
- **21 names have BOTH Kota and Kabupaten city forms** (bandung, bogor, depok is NOT among them — depok is only Kota; bandung, malang, semarang, tangerang, tegal, etc. have both).
- **Critical behavior confirmed:** `Jl. Aceh, Bandung` → "aceh" is a **road name**, resolved as nothing; it lands in `ambiguous` (unused evidence). `Jl. Kenangan, Depok, Yogyakarta` → Yogyakarta *does* resolve as a city candidate (Kota Yogyakarta 0.37) but the Depok subdistrict chain (0.47) wins.

**Why "city wins" must be gated on "no other evidence":**
- `Jl. Kenangan, Depok, Yogyakarta` has Yogyakarta as city-level evidence → priority for `depok` must be **ignored**; the address should resolve toward Yogyakarta (Depok is a *kecamatan* in Sleman, DI Yogyakarta).
- `Depok, Jawa Barat` has province evidence → today already resolves correctly (Kota Depok 0.82). Priority is not needed, and must not fight the province-pinned match.
- `Depok 16431` (postal code 16431 = Pancoran Mas, Kota Depok) → postal code is evidence → priority not needed; the postal match already wins (0.72).
- **But** `Jl. Aceh, Bandung` has NO resolved location evidence (aceh is just a road name) → priority **applies** → `bandung` must resolve as city.
- `Bandung` alone (no evidence at all) → priority applies → city.

So the rule is: **suppress SUBDISTRICT/DISTRICT entities for a priority name only when the address has no other resolved location evidence** (no province, city, district, subdistrict, or postal-code evidence besides the token itself).

---

## Task 1: `.gitignore` — ignore plan markdown files

**Objective:** Any `*.md` inside `plan/` is ignored by git; the `plan/` directory itself stays trackable.

**Step 1: Edit `.gitignore`** (done already)

```
plan/*.md
```

**Step 2: Verify**

```bash
git check-ignore plan/benchmark_prio_city_over_other_level_plan.md   # prints the path
```

**Step 3: Commit** (done as `2b9b5b1`)

---

## Task 2: Database — new `location_city_priority` table (with KOTA/KABUPATEN)

**Objective:** Store, per source, city names duplicated as subdistrict/district names, plus whether the preferred city form is Kota or Kabupaten.

**Files:**
- Modify: `db/location.sql` (add table + index)
- Modify: `internal/database/location.go` (add `FindAllCityPriority` returning name + type)

**Step 1: Add table to `db/location.sql`**

```sql
-- location_city_priority: normalized city names that ALSO exist as
-- subdistrict/district names in the same source, with the preferred
-- city form type ("KOTA" or "KABUPATEN") to use when the token is
-- resolved as a city with no other disambiguating evidence.
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

## Task 3: Seeder — populate `location_city_priority` with city_type

**Objective:** After seeding + hierarchy rebuild, insert every city-level normalized name that also appears at level 4 or 5, with `city_type` = `KOTA` if any city with that name is a "Kota X", else `KABUPATEN`.

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
    _, err := r.db.ExecContext(ctx, `
        INSERT OR IGNORE INTO location_city_priority (location_source_id, lowercase_normalized, city_type)
        SELECT DISTINCT c3.location_source_id, c3.lowercase_normalized,
               CASE WHEN MAX(CASE WHEN c3.name LIKE 'Kota %' THEN 1 ELSE 0 END) = 1
                    THEN 'KOTA' ELSE 'KABUPATEN' END
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

Note: city_type resolution — `KOTA` if **any** city row for that name is `Kota X`; else `KABUPATEN`. This matches the user's rule: "use kota unless there is no other than kab". (21 names have both → `KOTA`.)

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
sqlite3 db/location.db "SELECT lowercase_normalized, city_type FROM location_city_priority WHERE lowercase_normalized IN ('bandung','depok','surabaya','kupang','malang');"
# bandung -> KOTA, depok -> KOTA, surabaya -> KOTA, kupang -> KOTA, malang -> KOTA
```

---

## Task 4: Service — load priority set once into memory

**Objective:** Load `location_city_priority` into a `map[string]string{name: type}` once per source, following the `sync.Once` cache pattern.

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

**Step 3: Wire into `ensureEntitiesCachesLoaded`** in `internal/service/resolve.go`:

```go
if err := ensureCityPriorityLoaded(svc, ctx, sourceID); err != nil {
    return err
}
```

**Step 4: Verify build + small loader unit test**

```bash
go build ./...
go test ./internal/service/ -run TestLoadCityPriority -v
```

---

## Task 5: Core fix — city wins, but only with no other evidence

**Objective:** During entity resolution, when a token is in `cityPrioritySet`:
1. If the token is the **only** resolved location evidence → suppress its SUBDISTRICT/DISTRICT entities (city wins), and prefer the `city_type` city (KOTA over KABUPATEN) when ranking city candidates.
2. If there is **other** resolved evidence (province/city/district/subdistrict/postal code) → priority is ignored; existing behavior stands.
3. Explicit `kecamatan`/`kelurahan`/`desa` context → always keeps the SUBDISTRICT/DISTRICT entity (priority never overrides explicit admin context).

**Files:**
- Modify: `internal/service/resolve.go` (`ResolveEvidence` — apply suppression + city-type preference)
- Modify: `internal/service/builder.go` (`DiscoverCandidates` / `collectUniqueEntities` — city-type preference ranking)
- Modify: `internal/service/validate_helper.go` (explicit-context detection helper)

**Step 1: Detect explicit admin context (pre-normalization)**

In `ValidateAddressV1`, before normalization, scan the sanitized string for `kecamatan|kec\.?|kelurahan|kel\.?|desa` immediately preceding a priority name. Build a `map[string]bool` of tokens that must keep their lower-level entities:

```go
// regex: (kecamatan|kec\.?|kelurahan|kel\.?|desa)\s+(NAME)
// NAME = the priority token (normalized) found in the raw text
```

If the token follows an admin prefix, do NOT suppress its SUBDISTRICT/DISTRICT entities even though it's in the priority set.

**Step 2: Suppression in `ResolveEvidence`**

In `ResolveEvidence`, after resolving all evidence, determine whether the priority token is the only resolved location evidence:

```go
// gather all resolved entities across evidence
// for each evidence token whose resolved candidates include a priority name:
//   if the token is NOT explicitly kecamatan/kelurahan/desa'd
//   AND no OTHER evidence token resolved to any location entity
//   -> drop SUBDISTRICT and DISTRICT candidates for that token
//   -> keep CITY candidates only
```

Implementation sketch (in `ResolveEvidence` after the existing loop):

```go
priorityName := "" // normalized name of the priority token
for _, ev := range evidence {
    if svc.isExplicitAdminToken(sanitized, ev.Value) { continue }
    if _, ok := svc.cityPrioritySet[ev.Value]; ok {
        priorityName = ev.Value
        break
    }
}
if priorityName != "" {
    // count resolved evidence outside this token
    otherEvidence := 0
    for _, re := range resolved {
        if re.Evidence.Value == priorityName { continue }
        if len(re.Candidates) > 0 { otherEvidence++ }
    }
    if otherEvidence == 0 {
        // suppress lower levels for this token
        for i := range resolved {
            if resolved[i].Evidence.Value != priorityName { continue }
            filtered := resolved[i].Candidates[:0]
            for _, c := range resolved[i].Candidates {
                if c.Level == "CITY" { filtered = append(filtered, c) }
            }
            resolved[i].Candidates = filtered
        }
    }
}
```

**Step 3: City-type preference (KOTA over KABUPATEN)**

When multiple CITY candidates remain for the priority token and no other evidence disambiguates, prefer the city whose name matches the `city_type`:

- `bandung` → city_type `KOTA` → prefer `Kota Bandung` over `Kabupaten Bandung`.
- Names with only Kabupaten (e.g. `karanganyar`) → city_type `KABUPATEN` → `Kabupaten Karanganyar`.

Where to apply: in the confidence/ranking path. The cleanest spot is `DiscoverCandidates`/`buildFromCity` — when the priority applies, order city candidates with `city_type`-matching names first, and/or give them a small boost in `scoreConfidence` (e.g. a `WeightCityPriority` constant, suggested `0.02` — just enough to break ties without dominating province/postal evidence).

**Step 4: Unit tests** (new `internal/service/priority_test.go`)

```go
func TestCityPrioritySuppressesSubdistrict(t *testing.T) {
    // input "Bandung" with priority set {"bandung":"KOTA"}
    // assert resolved candidates for "bandung" contain CITY only, no SUBDISTRICT/DISTRICT
    // assert winner is Kota Bandung (not Kab. Bandung, not Tulungagung subdistrict)
}

func TestCityPriorityIgnoredWithOtherEvidence(t *testing.T) {
    // input "Jl. Kenangan, Depok, Yogyakarta"
    // Yogyakarta resolves as city evidence -> priority for depok is IGNORED
    // assert depok keeps its subdistrict/district candidates
}

func TestCityPriorityIgnoredWithPostalCode(t *testing.T) {
    // input "Depok 16431" -> postal code evidence -> priority ignored
    // assert resolves to Pancoran Mas, Kota Depok (postal path)
}

func TestCityPriorityIgnoredWithProvince(t *testing.T) {
    // input "Depok, Jawa Barat" -> province evidence -> priority ignored
    // assert resolves to Kota Depok (already correct today, must not regress)
}

func TestCityPriorityRespectsExplicitKecamatan(t *testing.T) {
    // input "kecamatan bandung" -> explicit context keeps SUBDISTRICT entities
}

func TestCityPriorityKabupatenOnly(t *testing.T) {
    // name with only Kabupaten form (e.g. karanganyar) -> city_type KABUPATEN
    // assert resolves to Kabupaten Karanganyar
}

func TestRoadNameNotTreatedAsEvidence(t *testing.T) {
    // input "Jl. Aceh, Bandung" -> "aceh" is a road name, resolves to nothing
    // priority applies -> bandung resolves as city (Kota Bandung)
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
- Challenge #2 (`city`) count should drop from **25**.
- `JL. GATOT SUBROTO NO.86 BANDUNG` → Kota Bandung (currently Tulungagung subdistrict).
- `Jl. Kenangan, Depok, Yogyakarta` → should NOT regress to Kota Depok; must resolve toward Yogyakarta (priority ignored due to Yogyakarta evidence).
- Exact match % should rise from **49.1%**.
- Check `page/benchmark.html` challenge section.

**Step 4: Regression check**

```bash
go test ./...
make test-api-smoke
```

---

## Acceptance Criteria

- [ ] `plan/*.md` ignored by git, `plan/` dir not ignored.
- [ ] `location_city_priority` exists, 234 rows, `city_type` populated (`KOTA` for bandung/depok/surabaya/kupang/malang).
- [ ] `Bandung` (no other evidence) → resolves to **Kota Bandung**.
- [ ] `JL. GATOT SUBROTO NO.86 BANDUNG` → Kota Bandung (challenge #2 example fixed).
- [ ] `Jl. Kenangan, Depok, Yogyakarta` → priority ignored (Yogyakarta evidence wins; no regression to Kota Depok).
- [ ] `Depok, Jawa Barat` and `Depok 16431` → unchanged (already correct).
- [ ] `Jl. Aceh, Bandung` → Kota Bandung (road name is not evidence).
- [ ] `kecamatan bandung` → subdistrict path preserved.
- [ ] Names with only Kabupaten form (e.g. `karanganyar`) → Kabupaten city.
- [ ] Benchmark challenge #2 < 25; exact-match % > 49.1%.
- [ ] All Go tests + k6 smoke pass.

## Risks / Tradeoffs / Open Questions

1. **Road names as non-evidence (verified):** `Jl. Aceh, Bandung` treats "aceh" as a road name (unresolved). This is why the "no other evidence" check must count *resolved* evidence only. Road-name handling is challenge #3's scope, not this plan's.
2. **KOTA preference is a tie-break, not an override:** when province evidence exists (`Depok, Jawa Barat`) the province already selects the right city; the priority only breaks ties among equal city candidates. Keep the boost small (`WeightCityPriority ≈ 0.02`) so it never overrides real evidence.
3. **Kabupaten-only names:** `city_type` handles them (`karanganyar` → KABUPATEN). Names with both forms default to KOTA per the user's rule.
4. **Multi-token names (25 of 234)** never match as words today (e.g. `tebing tinggi`), so they're inert in the priority set — no regression risk, just no benefit yet.
5. **Seeding dependency:** existing deployed DBs need re-seed (`bin/seeder --truncate` + seed) before the service picks up the table. Deployment must include it.
6. **What counts as "other evidence"?** Defined as any resolved entity outside the priority token: province, city, district, subdistrict, or postal-code evidence. Unresolved tokens (road names, noise) do NOT count. This is exactly what makes `Jl. Aceh, Bandung` apply priority while `Jl. Kenangan, Depok, Yogyakarta` does not.

## Verification commands (quick reference)

```bash
go build ./... && go test ./...
make benchmark && make benchmark-page
git check-ignore plan/benchmark_prio_city_over_other_level_plan.md
sqlite3 db/location.db "SELECT COUNT(*) FROM location_city_priority;"
sqlite3 db/location.db "SELECT lowercase_normalized, city_type FROM location_city_priority WHERE lowercase_normalized='bandung';"
```
