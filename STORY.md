# STORY.md

> The story of Address Quality: why it exists, what it does, how it was built, and where it is going.
>
> Sources: `ABOUT.md` (canonical technical facts) and `README.md` (vision and messaging). Where the two disagree, this document follows ABOUT.md on technical facts and notes discrepancies inline and in the appendix.

---

## Address Quality: the address intelligence layer for Indonesian addresses

Most software systems assume that addresses arrive clean, structured, and internally consistent. For Indonesian addresses, that assumption rarely holds.

Address Quality is an HTTP API that validates and normalizes Indonesian (Bahasa Indonesia) free-text addresses against a reference administrative hierarchy, and returns structured metadata: a confidence score, a quality status, resolved location components, and explainable evidence for every result.

It is positioned not as a geocoder or a logistics provider, but as an **address intelligence layer** — a validation step that runs *before* addresses reach downstream systems.

---

## The problem: bad data gets expensive late

End users commonly submit addresses containing abbreviations, aliases, landmarks, RT/RW information, typographical errors, incomplete administrative hierarchy, or informal location descriptions. These variations are natural for humans, but they introduce ambiguity for software that relies on deterministic matching.

Without a validation layer, applications forward raw addresses directly into geocoders, logistics providers, KYC platforms, and customer databases. Once poor-quality address data enters these systems, correcting it becomes significantly more expensive — resulting in failed deliveries, duplicate customer records, inaccurate administrative mapping, inconsistent analytics, and unnecessary manual verification.

Address Quality is designed to solve this problem before the address reaches downstream systems.

### Why Indonesian addresses are hard

Indonesian addresses combine official administrative regions with informal conventions that vary between cities and communities. A single address may contain road names, RT/RW information, landmarks, local abbreviations, postal codes, or administrative names written with different spellings.

A seemingly simple address hides real ambiguity:

```text
Jl. Merdeka No. 1
Bogor
```

"Bogor" alone could refer to **Kota Bogor** or **Kabupaten Bogor**. Meanwhile, all of the following describe the same administrative locations despite different text:

```text
Kab Bogor        Kabupaten Bogor
DIY              Daerah Istimewa Yogyakarta
IV Koto          4 Koto
Gg Mawar         Gang Mawar          
```

Administrative changes add further complexity: new regions are created, postal codes evolve, aliases emerge, and multiple administrative areas frequently share the same names. Without understanding the administrative hierarchy itself, software has very little context to determine whether an address is internally consistent.

---

## The vision: an intelligence layer, not a replacement

Geocoding APIs such as Google Maps or OpenStreetMap convert addresses into geographic coordinates. Logistics APIs generally assume an address was already validated before entering their systems. Neither is primarily responsible for evaluating whether an Indonesian address is complete, internally consistent, or administratively valid.

Address Quality is intended to **complement** these systems, not replace them. It sits between the raw address and the rest of the stack:

```text
Raw Indonesian Address
        │
        ▼
Address Intelligence Layer
  Normalize → Parse & Match → Validate Hierarchy → Resolve Ambiguity → Confidence & Evidence
        │
        ▼
Structured Address
        │
        ├──▶ Geocoder
        ├──▶ Logistics Platform
        ├──▶ CRM / KYC
        └──▶ Data Warehouse
```

Instead of returning only coordinates, it returns structured signals: normalized administrative components, confidence scores, ambiguity detection, candidate matches, and explainable validation evidence. Applications remain in control of how those signals are interpreted — accepting an address automatically, requesting user correction, triggering manual review, or continuing to geocoding.

The design philosophy is explainability: every result is traceable to the input evidence. Rather than producing only a confidence value, Address Quality explains *why* a result is reliable, which makes the output suitable both for automated decision making and for debugging and operational review.

---

## The product today

Address Quality is a single-purpose API with a deliberately small surface:

- **`POST /v1/validate`** — validate and resolve an Indonesian address.
- **`GET /health`** — health check (public, no auth).
- **`GET /swagger/*`** — Swagger UI (public, no auth).

Requests carry an address and an optional data source; responses include the request ID, address ID, quality status, confidence, matched/missing components, conflicts, ambiguity signals, resolution candidates, and metadata about the underlying data source.

### The administrative hierarchy

Validation runs against the Indonesian administrative hierarchy, sourced from a Kemendagri administrative dataset:

- **Province** (top level)
- **City / Regency** (kota and kabupaten)
- **District** (kecamatan)
- **Subdistrict** (kelurahan/desa) — carries the postal code
- **Postal code** — 5-digit

The hierarchy is encoded with the dot-separated Kemendagri `kode` (e.g. `11.01.01.2001`), and each subdistrict is mapped to its province/city/district uppers through a precomputed hierarchy table.

### Quality statuses

Every result is assigned one of five statuses:

- **VALID** — administrative components are consistent and complete.
- **INCOMPLETE** — a required component (province, city, or district) is missing.
- **AMBIGUOUS** — the top two candidates are too close to call.
- **CONFLICT** — internal inconsistency detected (e.g. hierarchy, postal-code mismatch).
- **UNKNOWN** — no administrative component could be matched.

### Confidence

Confidence is a score from 0 to 1, computed as a weighted sum of matched signals and clamped to the unit interval. Evidence weights include exact matches, hierarchical consistency, province/city/district/subdistrict matches, postal-prefix matching, and multi-evidence bonuses.

A short, illustrative example from the README:

```text
Province      ✔ Jawa Barat
City          ✔ Kota Bandung
District      ✔ Bandung Wetan
Subdistrict   ✔ Citarum
Postal Code   ✔ Match
```

In addition of scoring the address, the engine attaches the matched evidence behind each signal. It is the reasons that justify the result.

---

## How the engine works

Every request runs through a single deterministic validation pipeline. Matching is phrase/hash lookup against precomputed normalized names — there is no AI/ML model in the runtime engine.

The pipeline, in order:

1. **Sanitize** — HTML sanitization of the raw input.
2. **Normalize** — lowercase; strip administrative prefixes (`kabupaten`, `kab`, `kota`, `kecamatan`, `kec`, `kelurahan`, `kel`, `provinsi`, `prov`, `administrasi`, `kepulauan`); remove non-alpha punctuation; collapse whitespace; preserve 5-digit postal codes.
3. **Input validation** — required, and within the maximum length.
4. **Evidence extraction** — find a 5-digit postal code; classify remaining words as road-name or place-name evidence.
5. **Entity resolution** — map each piece of evidence to *all possible* administrative entities using longest-match over the phrase dictionary. Resolution never picks a winner; it collects possibilities.
6. **Candidate discovery** — build one candidate per unique entity at each level (province, city, district, subdistrict), top-down and any-level.
7. **Candidate deduplication** — merge candidates that resolve to the same location.
8. **Enrichment** — fill upper levels from the hierarchy and attach postal codes.
9. **Conclusion building** — attach matched evidence to each candidate.
10. **Evaluation** — validate hierarchy consistency, assess completeness, measure evidence coverage, detect conflicts, compute confidence, assign a quality status, and generate reasons.
11. **Ranking** — sort by confidence, then filled-level count, then fewer conflicts.
12. **Ambiguity check** — if the top two candidates are within 0.1 confidence and the top is VALID, the result becomes AMBIGUOUS.
13. **Postal-code fallback** — if the winner is empty, infer the full location from the postal-code database; if the winner lacks a postal code, fill it.
14. **Response** — build the formatted address (`subdistrict, district, city, province [postal]`) and serialize the structured response.

Because the engine evaluates administrative consistency across multiple hierarchy levels rather than a single string comparison, it can detect internally inconsistent addresses — and explain exactly which evidence supports (or fails to support) each interpretation.

---

## The build story

Address Quality was built over roughly four weeks (July–August 2026) by a single author, in a linear history of 130 commits with no release tags. Ten phases tell the story: an MVP API, API hardening, containerization and first CI, the data layer and seeder, caching and candidate sets, benchmarks/docs/licensing, a full evaluation-engine overhaul, Swagger documentation, the frontend (later moved to samaita.com), and production CI/CD polish.

A few chapters are worth telling in more detail.

### Shipping the whole administrative tree in a SQLite file

The validation engine is only as good as its reference data. The seeder parses roughly 176,000 lines of upstream MySQL dumps — the `wilayah` administrative dump and the `wilayah_kodepos` postal-code dump — and rebuilds them into a normalized SQLite hierarchy: **38 provinces, 514 cities/regencies, 7,285 districts, and 83,762 subdistricts**, joined with postal codes, with `lowercase_normalized` names for phrase matching and a precomputed hierarchy table for fast upper-level lookups.

This makes the reference dataset fully offline: no network dependency at validation time, no external service. A pure-Go SQLite driver keeps the binary CGO-free and statically linked.

### Two databases with two very different jobs

The project uses two separate SQLite files with distinct personalities:

- **`location.db`** — the immutable, read-only administrative hierarchy. Loaded once into in-memory maps and served from memory for the life of the process.
- **`address.db`** — the append-only request log, recording every validation request and its output.

Separating static reference data from mutable log data keeps the hierarchy safely read-only and the log trivially append-only.

### A scoring model you can unit-test

The confidence weights live in a single file and are asserted to exact values in table-driven tests — an exact-only match scores 0.10, a hierarchy-valid-plus-city match 0.15, a full pipeline 0.37, multi-evidence combinations 0.52/0.60. This calibration loop, visible in the git history as a "Scoring Issue Fix" saga, makes the scoring model comparable and testable.

### CI that builds but never deploys

CI builds and pushes one immutable container image to a registry — the backend API — and **never deploys**. Deployment is a manual, operator-run step on the production VPS: a script pulls the chosen image tag (including immutable `sha-` tags for reproducibility), recreates the container, waits for the health check, verifies the reverse proxy, and records the deployed tag to a release file for one-command rollback.

This security-oriented split means there is no CI-to-production path, and production changes are always a deliberate, auditable operator action.

### A frontend that was built, then moved on

A React/TypeScript frontend — marketing site, API documentation, and an interactive playground — was built in this repo with Vite, Tailwind, and a small component set, deliberately kept lightweight and AI-assisted ("vibe coded") so engineering effort stays focused on the validation engine. It was later removed from the repository and is now hosted on the main samaita.com site, keeping this repo purely about the API.

---

## Current state

Address Quality is in **Public Alpha** — under active development, already usable, but with no backward-compatibility guarantee before the first stable release. The validation algorithm, scoring model, and API responses are expected to keep evolving as new datasets, edge cases, and validation strategies are incorporated.

What is in place today:

- Administrative hierarchy validation (province → city → district → subdistrict → postal code)
- Alias and abbreviation normalization
- Candidate ranking and confidence scoring
- Explainable validation evidence
- Postal-code verification
- Indonesian-first parsing strategy
- Offline administrative database
- API-key authentication and per-IP rate limiting
- Go unit tests, k6 smoke/load tests, and an accuracy benchmark against a tagged ground-truth CSV

---

## What's next

The roadmap marks five items as done and five as open:

| Done | Open |
|---|---|
| Administrative hierarchy parser | Road-level validation |
| Candidate generation | OpenStreetMap integration |
| Confidence scoring | Google Maps fallback |
| Explainable evidence | Batch validation API |
| Postal code validation | Official SDKs |

Beyond the roadmap, the known limitations worth knowing before adopting the API:

- **Road-level validation is not implemented** — road names are extracted as evidence but not resolved.
- **No geocoding integration** — the API validates; it does not produce coordinates.
- **No batch validation API** and **no official SDKs** yet.
- **The shared API key is the only authentication mechanism** — all valid keys are equal, with no scoping.
- **Production TLS is not configured in the repository** — the host nginx example listens on plain HTTP.
- **The rate limiter keys on the connection's remote address**, which is proxy-unaware in the current implementation.

The project also ships under the Business Source License 1.1, with a separate Commercial License for production and commercial use, and its license states that future major versions may be released under different terms.

---

## The road ahead

Address Quality exists to reduce the cost of poor-quality Indonesian address data reaching downstream systems. Its bet is that the right architecture for that problem is a deterministic, explainable validation engine running against an authoritative, offline administrative hierarchy — one that tells you not just *how confident* a result is, but *why*.

The validation engine is the product. Everything else — the deployment tooling, the docs, the playground, even the frontend that now lives on samaita.com — is in service of that. As the dataset grows, edge cases accumulate, and real-world feedback arrives, the engine is designed to keep improving without ever asking the caller to trust a black box.

---

## Appendix: source-consistency notes

The two source documents (ABOUT.md, the canonical technical reference, and README.md, the project's public messaging) mostly agree. The notable discrepancies are listed here; per the ground rules, technical facts follow ABOUT.md.

1. **Administrative dataset citation.** README.md cites "Keputusan Menteri Dalam Negeri No300.2.2-2430 2025" as the basis of the administrative hierarchy. The seeded database and the source dump header state "No 300.2.2-2138". The authoritative number is unresolved from repository evidence.
2. **Confidence scale in the example.** The README's illustrative output shows "Confidence 97"; the API actually returns a confidence score clamped to the range 0–1 and rounded to four decimals. The README figure is a display convention, not the on-the-wire format.
3. **Batch normalization as a use case.** README lists "Batch address normalization" as a target use case, but there is no batch validation API yet — it is an open roadmap item.
4. **Swagger copy-paste drift.** Some API documentation artifacts (the Swagger description and the server `@description` annotation) read "Thai address validation and resolution API" — a copy-paste artifact; the product validates Indonesian addresses.
5. **Scope of active discovery strategies.** README presents alias/abbreviation normalization as a current feature; the `alias` and `postal` discovery strategies are declared in code but not yet exercised at runtime — only `top_down` and `any_level` are active.

---

*Derived from `ABOUT.md` and `README.md`. Technical facts follow ABOUT.md; vision and messaging follow README.md.*
