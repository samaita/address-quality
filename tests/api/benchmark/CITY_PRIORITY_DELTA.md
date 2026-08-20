# City Priority Fix — Benchmark Delta Analysis (2026-08-20)

## Baseline vs fix (benchmark v1_0000 vs v1_0001, 106 records)

| Metric | Baseline | After fix | Delta |
|---|---|---|---|
| Exact match (all 4 levels) | 52 (49.1%) | 52 (49.1%) | 0 |
| Province correct | 87 (82.1%) | 92 (86.8%) | **+5** |
| City correct | 81 (76.4%) | 85 (80.2%) | **+4** |
| District correct | 76 (71.7%) | 76 (71.7%) | 0 |
| Subdistrict correct | 55 (51.9%) | 55 (51.9%) | 0 |
| Challenge #2 (city wrong) | 25 | 21 | **-4** |
| Challenge #3 (road) | 22 | 18 | -4 |

## Per-record delta

- **FIXED (correct->correct level gains): 5 records** gained province, 4 gained city.
  All 5 were the same class: `...BANDUNG` addresses that previously resolved to
  "Bandung" subdistrict in Kab. Tulungagung / Kab. Serang (wrong city, wrong province).
  Now resolve to Kota Bandung / Kabupaten Bandung (correct province + city):
  - JL. GATOT SUBROTO NO.86 BANDUNG -> Kota Bandung (was Tulungagung)
  - JL. PETA NO.176 BANDUNG -> Kota Bandung
  - SANTIKA HOTEL BANDUNG, JALAN SUMATERA NO.52-54 -> Kota Bandung
  - THE PAPANDAYAN HOTEL, JL. GATOT SUBROTO NO.83 BANDUNG -> Kota Bandung
  - JL. Terusan Ir. Sutami No.62, Bandung -> Kabupaten Bandung (city still wrong: should be Kota)
- **REGRESSED: 0 records.** No previously-correct record became wrong.

## Level-matched comparison (the real improvement)

Exact-match (4/4 levels) is the "first-class" outcome, but it hides partial
improvement. Comparing matched-levels per record:

| Levels matched | BEFORE | AFTER |
|---|---|---|
| 0 levels | 18 | **13** |
| 1 level | 3 | 4 |
| 2 levels | 11 | **15** |
| 3 levels | 22 | 22 |
| 4 levels (exact) | 52 | 52 |
| **Average levels matched** | **2.821** | **2.906** |

Per-record shift: **+2 levels: 4 records, +1 level: 1 record, +0: 100.**
Improved: 5, worsened: 0, unchanged: 100.

The 5 improved records went from 0/4 levels (wrong province, city, everything —
Tulungagung/Serang subdistrict chains) to 2/4 (correct province + city: Kota
Bandung, Jawa Barat). Road-type addresses: avg levels 2.850 -> 2.940.

**Verdict: real improvement, invisible in the strict 4/4 headline.**
The fix moved 5 records from "completely wrong" to "correct city+province",
which matters for logistics (right city is the actionable unit). The 4/4 metric
did not move because those records still miss district/subdistrict (road-level,
no street data -> challenge #3).

## Why exact-match % did not move

Exact match requires province+city+district+subdistrict all correct. The 5 fixed
records now get province+city right but still miss district/subdistrict (road-based
addresses, no street data -> challenge #3). The fix traded "wrong city + wrong
everything" for "right city + missing lower levels" — a real accuracy gain for
city-level use cases, invisible in the strict all-4-levels metric.

## Remaining challenge #2 (21 records)

- 18 of 21 are road-like addresses (challenge #3 overlap): the road name token
  resolves as a subdistrict elsewhere and OUTSCORES the city token in ranking
  (e.g. "JL. CIMANUK ... BANDUNG" -> Cimanuk/Pandeglang 0.43 > Kota Bandung 0.37;
  "JL. FAJAR 78 B KPAD GERLONG" -> Fajar/Tapanuli Tengah).
  The city priority correctly stops road tokens from BLOCKING the fix, but does not
  stop them from WINNING the ranking. That is a scoring/road-challenge issue.
- 3 of 21 non-road:
  - "23 PASKAL MALL NO.#L1-60", "MALL PASKAL L1-10, 10T, L12" — no location tokens, empty result
  - "Sukarasa, Sukasari, Jawa Barat" — district-level ambiguity (sukasari is not a
    city-overlap name; subdistrict in Purwakarta vs Bandung tie at 0.78)

## Verdict

- The city-priority fix works as designed: city names that overlap subdistrict/district
  now resolve as city (Kota preferred) when no other evidence pins the location.
- Zero regressions. City-level accuracy +4, province +5.
- The exact-match % (49.1%) is unchanged because the metric is stricter than what
  this fix targets; the residual misses are road-name-vs-city scoring (challenge #3)
  and district-level ambiguity — separate follow-up plans.
- Known open item (pre-existing): "Jl. Kenangan, Depok, Yogyakarta" correctly ignores
  priority (depok not suppressed) but still resolves to Kota Depok 0.47 vs Yogyakarta
  0.37 — multi-city scoring gap, separate follow-up.

## Scripts (this dir)

- scratch_delta.py  — fixed/regressed/level-movement between 0000 and 0001
- scratch_detail.py — per-record candidate comparison for the 5 fixed records
- scratch_city21.py  — breakdown of remaining challenge #2 records
