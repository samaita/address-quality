# Benchmark Prio: City Over Other (City-Name Collision Fix)

> **Goal:** When an address contains a token that is both a city name and a subdistrict/district name, resolve it as the **city** — unless other `place_name` evidence already pins the location. When the city is chosen and both Kota and Kabupaten forms exist, prefer **Kota** unless only Kabupaten exists.

**Architecture:** No new table, no new lookup. The collision set is derived from the **existing caches** (province/city/district/subdistrict, all keyed by `normalizer.Normalize(name)` under `{sourceID}:{normalized}`) at load time, and held in a small `map[string]string` (normalized name → `"KOTA"`/`"KABUPATEN"`). Suppression happens during entity resolution, gated on **other place-name evidence**.

**Tech Stack:** Go (service layer), existing normalizer + caches, SQLite (`location.db`), k6 benchmark harness.

**Current state (2026-08-20 benchmark, 106 records):** 49.1% exact matches (52/106). Challenge #2 = 25 records (23.6%).

**Root cause (verified against `db/location.db` + live API):**

- `bandung` = 2 cities (Kab. 32.04, Kota 32.73) + 2 districts + **15 subdistricts** (e.g. `35.04.17.2008` = Bandung subdist. in Kab. Tulungagung, pc 66274).
- `JL. GATOT SUBROTO NO.86 BANDUNG` → winner `Bandung/Bandung/Kab. Tulungagung/Jawa Timur 66274` (0.43, `match_district, match_subdistrict`) — the enriched subdistrict chain outscores the bare city candidate (0.37).
- **234** normalized city names also exist at level 4/5; **21** have both Kota and Kabupaten city forms.
- Verified: `Jl. Kenangan, Depok, Yogyakarta` → "yogyakarta" resolves as a city candidate (Kota Yogyakarta 0.37) but the Depok subdistrict chain (0.47) wins. `Jl. Aceh, Bandung` → "aceh" is a **road name**, resolves to nothing (sits in `ambiguous`).

---

## Design summary (user directives)

1. **No new table / no new in-memory lookup structure.** The priority set is computed from the **existing caches** using the **existing `normalizer.Normalize`** function — the same key that already indexes every cache (`{sourceID}:{normalized}`).
2. **Other-evidence detection = `place_name` evidence only.** A token that resolves in the DB is evidence (that's the existing logic). If any **other** `place_name` token resolves to a location entity, the city priority for the ambiguous token is **ignored**. Postal codes and road names do not count as place-name evidence:
   - Postal code — separate evidence type (`postal_code`); per user's earlier directive "Postal code is also evidence" it gates priority too (a resolved postal code pins the location, so priority is unnecessary and must not fight it).
   - Road names (`road_name` type) resolve to nothing today (verified) — they cannot pin a location, so they never block priority.
3. **Kota unless no Kab.** Derived at load time from existing city cache: if any city row for the name is `Kota X` → `KOTA`, else `KABUPATEN`.

---

## Task 1: `.gitignore` — un-ignore plan markdown files

**Objective:** `plan/*.md` must be committed for review; remove the ignore line.

**Step 1: Remove `plan/*.md` from `.gitignore`** (done — PR #1 branch)

**Step 2: Verify**

```bash
git check-ignore plan/benchmark_prio_city_over_other_level_plan.md   # should print nothing (exit 1)
```

---

## Task 2: Service — derive priority set from existing caches

**Objective:** Build `cityPrioritySet map[string]string` (normalized name → city type) once, from the already-loaded province/city/district/subdistrict caches. No DB change.

**Files:**
- Modify: `internal/service/service.go` (fields)
- Modify: `internal/service/validate_helper.go` (build in `loadPhraseDict` / a new step)

**Step 1: Add field to `Service`**

```go
// normalized city name -> "KOTA" | "KABUPATEN" for names that ALSO exist
// as subdistrict/district; empty map when none
cityPrioritySet map[string]string
```

No new `sync.Once` needed — computed inside the existing `loadPhraseDict` (which already runs once under `phraseDictOnce`).

**Step 2: Compute the set in `loadPhraseDict`**

The caches are already loaded by the time `loadPhraseDict` runs (`ensureEntitiesCachesLoaded` → provinces, cities, districts, subdistricts → then `ensurePhraseDictLoaded`). Add, at the start of `loadPhraseDict`:

```go
// city names that also exist as subdistrict/district
cityNameSet := make(map[string]struct{})
for _, entries := range svc.subDistrictCache {
    for _, e := range entries {
        cityNameSet[e.Name] = struct{}{}
    }
}
for _, entries := range svc.districtCache {
    for _, e := range entries {
        cityNameSet[e.Name] = struct{}{}
    }
}

// prefer KOTA if any city row for the name is "Kota X", else KABUPATEN
svc.cityPrioritySet = make(map[string]string)
for _, entries := range svc.cityCache {
    for _, e := range entries {
        if _, dup := cityNameSet[e.Name]; !dup {
            continue
        }
        if strings.HasPrefix(e.Name, "kota ") {
            svc.cityPrioritySet[e.Name] = "KOTA"
        } else if _, seen := svc.cityPrioritySet[e.Name]; !seen {
            svc.cityPrioritySet[e.Name] = "KABUPATEN"
        }
    }
}
```

Note: `e.Name` is the **normalized** name (`normalizer.Normalize(r.Name)` at load time) — same key used everywhere, so no re-normalization at request time. `Kota`/`Kabupaten` prefixes are stripped by the normalizer, so `e.Name` is the bare name (e.g. `bandung`); the `strings.HasPrefix(e.Name, "kota ")` check is a safety net (multi-word names like `kota batu` normalize to `batu`, so the bare name check governs). The loop over city entries sets `KOTA` when any city row is a Kota, and only falls back to `KABUPATEN` when none is — matching "use kota unless there is no other than kab".

**Step 3: Verify build + unit test**

```go
func TestCityPriorityDerivedFromCaches(t *testing.T) {
    // fake caches: city "bandung" (Kota), subdistrict "bandung"; city "karanganyar" (Kab), subdistrict "karanganyar"
    // assert cityPrioritySet["bandung"] == "KOTA" && cityPrioritySet["karanganyar"] == "KABUPATEN"
}
```

```bash
go build ./...
go test ./internal/service/ -run TestCityPriorityDerivedFromCaches -v
```

---

## Task 3: Core fix — city wins, gated on other place-name evidence

**Objective:** During `ResolveEvidence`, for a priority token, suppress SUBDISTRICT/DISTRICT candidates unless another place-name (or postal-code) evidence resolved to a location.

**Files:**
- Modify: `internal/service/resolve.go` (`ResolveEvidence`)
- Modify: `internal/service/validate_helper.go` (evidence-type helper)

**Step 1: Define "other evidence" check**

In `ResolveEvidence`, after the resolution loop, detect:

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
        // postal code is also evidence (user directive): if it resolved, it pins the location
        if re.Evidence.Type == model.EvidencePostalCode && len(re.Candidates) > 0 {
            return true
        }
        // road names never resolve today -> never count as evidence
    }
    return false
}
```

Note: a `postal_code` evidence *value* is a 5-digit number; its candidates are subdistricts from `resolvePostalCodeEntity`. If it resolved, the location is pinned — priority is unnecessary. If it didn't resolve (bad code), it can't pin anything, so it doesn't block priority.

**Step 2: Apply suppression in `ResolveEvidence`**

After the existing `for _, ev := range evidence` loop:

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

This is the minimal change: the suppression only kicks in when the token is the **sole resolved evidence**; otherwise the existing candidate-building, enrichment, and scoring run unchanged.

**Step 3: Explicit `kecamatan`/`kelurahan`/`desa` context**

The normalizer strips admin prefixes (`kecamatan`, `kelurahan`, `desa` are in `adminSet`), so "kecamatan bandung" and "bandung" look identical post-normalization. To honor explicit context, detect the admin prefix **before** normalization in `ValidateAddressV1`:

```go
// on the sanitized string, find priority tokens immediately preceded by
// (kecamatan|kec\.?|kelurahan|kel\.?|desa)
// collect them into explicitSubdistrictTokens map[string]bool
```

Then in the suppression step, skip tokens in `explicitSubdistrictTokens` (keep their SUBDISTRICT/DISTRICT entities).

**Step 4: Unit tests** (new `internal/service/priority_test.go`)

```go
func TestCityPrioritySuppressesSubdistrict(t *testing.T) {
    // "Bandung" alone; priority {"bandung":"KOTA"}
    // assert candidates for "bandung" contain CITY only; winner Kota Bandung
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

## Task 4: Benchmark — measure the improvement

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

## Acceptance Criteria

- [ ] No new DB table, no new loader — priority set derived from existing caches via `normalizer.Normalize` keys.
- [ ] `Bandung` alone → **Kota Bandung**.
- [ ] `JL. GATOT SUBROTO NO.86 BANDUNG` → Kota Bandung (challenge #2 example fixed).
- [ ] `Jl. Kenangan, Depok, Yogyakarta` → priority ignored (Yogyakarta place-name evidence wins).
- [ ] `Depok, Jawa Barat` / `Depok 16431` → unchanged.
- [ ] `Jl. Aceh, Bandung` → Kota Bandung (road name is not evidence).
- [ ] `kecamatan bandung` → subdistrict path preserved.
- [ ] `karanganyar` → Kabupaten (only Kab form).
- [ ] Benchmark challenge #2 < 25; exact-match % > 49.1%.
- [ ] All Go tests + k6 smoke pass.

## Risks / Tradeoffs / Open Questions

1. **Road names are non-evidence (verified):** `Jl. Aceh, Bandung` leaves "aceh" unresolved, so priority applies. This is correct per the user's directive (road names can't pin a location today).
2. **Postal code counts as evidence** (user directive): a *resolved* postal code blocks priority — it already pins the location better than the priority heuristic. An *unresolved* postal code (typo'd) does not block.
3. **KOTA is a tie-break, not an override:** priority only kicks in with no other evidence; among equal city candidates the `city_type` preference picks Kota when present. It never overrides province/postal evidence.
4. **Multi-token names (25 of 234, e.g. `tebing tinggi`)** never match as single words today — inert in the priority set; no regression risk, no benefit yet.
5. **No deployment/seed step needed** (unlike the previous design) — the fix is purely in-memory from data already loaded. `go build` + restart is the whole deploy.

## Verification commands (quick reference)

```bash
go build ./... && go test ./...
make benchmark && make benchmark-page
git check-ignore plan/benchmark_prio_city_over_other_level_plan.md   # expect nothing
```
