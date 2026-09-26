# Address Quality performance: current state

Audit scope: repository at `f4c5be8` (main) plus the audit/contract request. This document records code and checked-in test assets, not benchmark measurements. No application code, validation behavior, or infrastructure was changed for this audit.

## Summary

The service is a Go 1.26 / Echo HTTP API. `POST /v1/validate` resolves addresses against local in-memory indexes loaded from mandatory SQLite location data. An optional PostgreSQL similarity-query helper exists, but repository search shows no call site in the V1 validation path at this revision. Request persistence, when enabled, is asynchronous through a bounded queue into the separate address SQLite DB. The endpoint is behind body-size, API-key, rate-limit, and request-ID middleware.

There is already an Echo middleware that logs per-request method, URI, status, body sizes, request ID, remote IP, host, and elapsed duration. This is event logging, not an aggregate metrics source: it cannot reliably supply percentiles, throughput windows, concurrency, or resource series. The existing k6 load script runs five hard-coded addresses under a fixed 10s ramp / 30s hold / 10s ramp-down; it has checks, thresholds, and a JSON summary. Existing accuracy benchmark code instead makes sequential Node HTTP requests over a CSV and writes per-address correctness results; it does not measure performance. The benchmark HTML generator displays k6 load summary data as an optional section, but it does not provide a unified performance artifact or report.

## Repository findings

### Request and validation flow

1. `internal/router/router.go:Setup` creates Echo, configures read/write timeouts, installs logging, recovery, and CORS middleware, exposes `/health` and Swagger, and registers `/v1/validate` behind body-limit, API-key, rate-limit, and request-ID middleware.
2. `internal/handler/handler.go:HandleAddressRequest` binds `model.AddressRequest`, maps binding/validation errors to 400, calls `Service.ValidateAddress`, maps storage/internal errors to 500, and returns JSON 200 on success.
3. `internal/service/service.go:ValidateAddress` delegates to `ValidateAddressV1`.
4. `internal/service/validate_v1.go:ValidateAddressV1` validates input; sanitizes and normalizes it; finds the requested/default location source; lazily loads the province/city/district/subdistrict caches, hierarchy, and city-priority data; extracts evidence; detects road-context tokens; resolves evidence against phrase dictionaries; discovers candidates using top-down and any-level strategies; deduplicates and enriches candidates; builds conclusions; performs one contextual recovery pass and may rebuild candidates; evaluates and sorts candidates; applies optional postal-code lookup; assembles the response; and optionally submits an address record to the store queue.
5. Candidate enrichment, hierarchy lookup, phrase resolution, scoring, and fuzzy contextual comparison are primarily in-memory after the lazy cache load. `internal/service/fuzzy_context.go` searches candidate neighborhoods in memory. The optional `FindSimilarLocations` PostgreSQL query exists in `internal/database/location.go`, but repository search shows no call site in the V1 validation path at this revision. Do not attribute normal V1 request time to it without implementation changes proving otherwise.
6. Postal-code fallback may query SQLite through `LocationRepository.FindByPostalCode`; that function can issue additional parent-name lookups for each matching postal result.
7. Optional address storage uses `queue.Queue` (size 1000, two workers, five-second job deadline) and `InsertAddressRequest`. Queue rejection and write failure are logged; storage work does not synchronously gate a successful validation response. `ENABLE_STORE_REQUEST` defaults to false.

### Persistence and external calls

- Mandatory local persistence: address SQLite DB (storage/cache features) and location SQLite DB. `database.NewLocationDB` uses `modernc.org/sqlite`; connection maximum comes from `DB_MAX_OPEN_CONNS` (default 10). Runtime validation eagerly loads location data once using `sync.Once` caches, then serves most name/hierarchy resolution from memory.
- Optional PostgreSQL (`POSTGRES_DSN`) is opened at startup when configured, is reported by health checks, and provides a trigram similarity method. Its configured presence does not prove that V1 validation invokes the method.
- Google Maps is the external dependency of V0 geocoding (`/v0/validate`), not normal V1 address validation. It has an SQLite cache, a serialized check/fetch/cache section, mock mode defaulting to true, and a ten-second HTTP client timeout. Keep V0/external latency out of V1 baseline scenarios unless a separate V0 experiment is explicitly defined.
- No external network lookup is performed by the inspected V1 path.

### HTTP, logging, existing metrics and tracing

- HTTP framework: Echo v4. `cmd/server/main.go` binds to `:<PORT>` (default 7300), sets signal-aware shutdown, and requests a 15-second shutdown window. `router.Setup` configures `READ_TIMEOUT` (5 seconds default) and `WRITE_TIMEOUT` (10 seconds default).
- `internal/logger/logger.go:EchoMiddleware` times the downstream Echo chain with `time.Now`, then writes a structured zerolog event containing request ID, method, raw URI, client IP, host, status, bytes in/out, and `latency`. It logs 4xx as warnings and 5xx as errors. It neither records a histogram nor aggregates events. URI, IP, host, and request ID are unsuitable metric dimensions; raw address and API-key values must never be emitted as metric labels.
- Validation emits debug/error messages, and DB failures log operation plus `input`; some paths include request contents or tokens. Do not reuse those log fields as metrics, and review existing sensitive logging separately from this measurement contract.
- Repository search found no Prometheus, OpenTelemetry, `expvar`, pprof, runtime stats, or tracing instrumentation in application Go source. Existing request-duration logging is not metrics/tracing.

### Existing performance and benchmark assets

- `tests/api/smoke-test.js`: one VU and one iteration, exercising health, successful V1 validation, sanitization, invalid-empty input, and malformed body. It includes checks and writes a k6 JSON summary via `tests/api/summary.js`.
- `tests/api/load-test.js`: fixed five Indonesian example addresses, cyclic selection, stages of 10 seconds ramp-up, 30 seconds at `K6_VUS` (default 10), and 10 seconds ramp-down. It checks for HTTP 200/request ID, has `http_req_failed < 1%` and `checks == 100%` thresholds, and emits avg/min/med/max/p90/p95/p99 trend stats. It skips status 429 from a custom `success_duration` Trend but still fails its HTTP-200 check. This is a useful starter smoke/load scenario, not a representative traffic model. It has no scenario-specific request tags or realistic dataset sampling.
- `tests/api/summary.js` serializes k6 metrics, thresholds, checks, options, and run duration to JSON only when `RESULT_JSON` is provided. The run wrapper determines artifact output; stdout remains a human summary.
- `tests/api/benchmark-test.js` reads CSV (Make target supplies `tests/api/cases/address-tagged.csv`), makes requests sequentially using Node's `http` client, stores each full V1 response plus ground-truth comparison in a JSON array, and reports success/failure counts. It records neither client timing nor per-request HTTP status when writing rows and is an accuracy corpus runner, not a load generator. Latest checked-in V1 files have 106 records each; these committed results date from 2026-08-20.
- Available CSV case files counted at audit time: `address-tagged.csv` 61 rows, `address.csv` 109 rows, `example.csv` 4 rows. `address.csv` is gitignored/private per `tests/api/cases/README.md`; the tagged file is also ignored, so do not publish raw rows or copy them into public artifacts. The checked-in 106-result benchmark set is historical and not necessarily the same as the currently ignored inputs.
- `tests/api/page/build.js` reads the latest V1 accuracy JSON and latest load-test summary JSON, appends/replaces a `metadata.json` entry keyed by release/build label, computes accuracy, and embeds a performance summary if present. The public page trims most raw inputs; `full-benchmark.html` contains all raw addresses, is gitignored, and is explicitly local/private. `release.json` allows optional runner/server descriptors, but those are manually supplied rather than measured. The current report is an accuracy benchmark page with a limited optional k6 section, not the future performance audit report.
- There are no other checked-in k6 scripts at the repo root or in `tests/api` beyond smoke/load. The repo has performance-adjacent instructions but no dedicated performance audit/contract docs before this task.

### Deployment and runtime constraints

- `Dockerfile` builds a static Linux Go binary (`CGO_ENABLED=0`) using Go 1.26 Alpine, then runs as a non-root user in Alpine 3.21. Runtime image exposes 7300 and uses a wget health check.
- Production compose in `deploy/docker-compose.prod.yml` maps port 7300, mounts `/etc/address-quality/db`, restarts unless stopped, and declares 512 MiB / 1 CPU. `deploy.sh` deploys via Podman Compose. These are configured limits, not evidence of the host's available CPU/memory or the actual target's workload.
- Local audit environment: Linux amd64, Go 1.26.6, Node 22.23.2, four visible CPUs, `k6` present on PATH. Host reported memory is not a benchmark target specification. No performance workload was run for this audit; therefore there are no result values to report.

## Current-state measurement matrix

`AVAILABLE` means the source can produce that measurement now; `PARTIAL` means it exists only in a limited/non-authoritative form. Statuses describe capability at this revision, not whether a baseline dataset exists.

| Measurement | Status | Actual source / gap | Needed for controlled baseline? |
|---|---|---|---|
| Request count | PARTIAL | k6 summary `http_reqs` for k6 runs; sequential accuracy runner counts succeeded/failed in console but writes no timing/status summary | REQUIRED from k6 |
| HTTP request latency | PARTIAL | Echo logs one request-chain duration; k6 records client HTTP trends in its JSON summary | REQUIRED from k6; app log is not aggregate |
| Latency distribution | PARTIAL | k6 built-in `http_req_duration` trend is exported as aggregates; app only logs samples | REQUIRED at k6 |
| p50 | PARTIAL | k6 default trend has median (`med`); no contract names/validates p50 specifically | REQUIRED, explicitly derive p50 from the k6 trend |
| p90 | PARTIAL | k6 options request p90; summary serializes aggregate values | REQUIRED |
| p95 | PARTIAL | k6 options request p95; summary serializes aggregate values | REQUIRED |
| p99 | PARTIAL | k6 options request p99; summary serializes aggregate values | REQUIRED |
| Throughput / requests per second | PARTIAL | k6 `http_reqs` rate/value in summary and run duration; no curated benchmark artifact/report field | REQUIRED from k6, preserving unit/interval |
| Error rate | PARTIAL | k6 `http_req_failed` and check rates / thresholds; accuracy runner marks non-200/parse/network outcomes as failures but no rate artifact | REQUIRED from k6; keep transport/HTTP/check failures distinguishable |
| HTTP status distribution | MISSING | k6 has status in response context but no explicit bounded per-status summary in checked-in summary schema; app logger stores status per event only | REQUIRED for audit diagnostics; status code is bounded and suitable as a dimension |
| Concurrency | PARTIAL | k6 options contain VUs/stages; summary stores options and `vus` metric where applicable, but no explicit achieved concurrency series | REQUIRED configured VUs/stages; optional achieved VUs metric |
| CPU | MISSING | no process or container sampler in app/tests | OPTIONAL diagnostic; mark NOT MEASURED unless an external sampler is specified and host identity captured |
| Memory | MISSING | no process/container resource sampler | OPTIONAL diagnostic; do not substitute host MemTotal for process memory |
| Go runtime metrics | MISSING | no endpoint/exporter/snapshot code; stdlib could sample explicitly | OPTIONAL diagnostic; basic snapshot only if reliable, not required for first latency baseline |
| GC metrics | MISSING | no `runtime.MemStats` / runtime metrics collection | OPTIONAL diagnostic |
| Goroutine count | MISSING | no runtime sample/export | OPTIONAL diagnostic |
| Database/persistence latency | MISSING | DB methods do not time successful calls; DB errors only are logged. V1 normal lookups use startup-loaded caches; optional postal lookup queries SQLite. Async request-store time is not part of response completion | OPTIONAL for baseline; stage timing is more useful first. Add only if there is an active database diagnostic question |
| External dependency latency | NOT_APPLICABLE | V1 path has no external lookup; Google Maps belongs to V0 and is outside this baseline | NOT_REQUIRED for V1 baseline; separate V0 experiment only if requested |
| Internal Address Quality stage latency | MISSING | only whole-handler chain is logged; no stage timers/histograms | REQUIRED diagnostically for server-side decomposition, but not to establish client HTTP baseline |

## Actual validation boundaries and measurement decision

| Boundary in current code | Diagnostic question | Regression value / cost | Contract decision |
|---|---|---|---|
| Handler request processing (`EchoMiddleware` around route) | How long does the server spend across the middleware/handler path, versus k6 client-observed time? | Useful server/client decomposition; a single monotonic duration is cheap. Existing log duration is sampled per request, not aggregatable | Retain existing log; add a bounded aggregate server request measurement for benchmark diagnostics only if implementation offers scrape/snapshot without a public debug API. Not the source for client percentiles |
| Input sanitize + normalize (`ValidateAddressV1`) | Is text preparation the regression source? | Low collection cost, but likely small and not independently actionable until data shows it is material | Do not add to baseline contract; no separate metric initially |
| Source lookup + lazy cache initialization (`FindSourceByCode`, `ensureEntitiesCachesLoaded`) | Are first-request cold initialization and source selection dominating? | Could explain cold-vs-warm effects; cache initialization is once per service/source and materially changes request cost | REQUIRED as workload methodology: distinguish cold start/warmup from steady-state, but do not add per-call source/cache metric by default. If needed, instrument a single cache-load duration separately, not each cache query |
| Evidence extraction + exact phrase resolution (`ExtractEvidence`, `ResolveEvidence`, phrase matching) | Is evidence matching regressing? | Meaningful logical boundary; cheap timer around the combined stage. Fine-grained token/function timings are noisy and high-volume | Candidate optional stage timer; recommend one combined `evidence_resolution` duration if evidence is required from the first diagnostic baseline, bounded labels only |
| Candidate discovery, deduplication, enrichment, conclusions | Is hierarchy candidate construction regressing? | Major stage using candidates and in-memory hierarchy; a single timer and candidate count (distribution, not per-candidate labels) can help | REQUIRED stage timing as one combined `candidate_build` boundary; do not time every helper |
| Contextual recovery and conditional rebuild (`RecoverContextualEvidence`) | Is the second pass expensive, how often is it active? | Distinct algorithmic phase and potentially input-sensitive; timer plus boolean/count outcome is useful, but high cardinality inputs are not | REQUIRED as `contextual_recovery` duration; a low-cardinality outcome (attempted/recovered yes/no) is optional. Never label by address/token/candidate ID |
| Candidate scoring and ordering (`EvaluateCandidate`, sort) | Is per-candidate evaluation/ranking driving time? | Can diagnose scoring work but timing every candidate creates excess observations and overhead | OPTIONAL one combined `candidate_evaluation` duration for the full candidate loop plus sort; candidate count can accompany as numeric observation |
| Postal-code lookup (`resolveLocationByPostalCode`) | Is DB fallback responsible for latency? | Conditional SQLite query; diagnostic and cheap to time only when invoked. Not representative of all requests | OPTIONAL lookup duration and invoked/outcome count, with no postal-code labels |
| Async record storage | Is queue persistence delayed/failing? | Does not gate the request success path; relevant to queue health, not client latency unless saturation/rejection correlates | Exclude DB write latency from response-stage baseline. Optional future queue counters/depth if store mode is benchmarked; no address labels |
| External lookups | What is V1 upstream dependency time? | None in V1. Google Maps is V0 and external/cache behavior would change the question | Reject for V1 metric set |

Instrumentation should time stable coarse boundaries only. Do not expose every helper-function timer. For histogram labels, use fixed stage and bounded outcome dimensions; never labels for URI with parameters, request IDs, source codes if arbitrary, IP, raw/normalized address, evidence values, postal codes, candidate IDs, API keys, or error strings.

## Workload review and recommendations

| Scenario | Recommended source | Scope and caveat |
|---|---|---|
| Smoke | Existing `smoke-test.js` | Keep one iteration and functional checks. Run only against explicitly selected non-production/local target by default; do not interpret one iteration as a latency baseline. |
| Baseline | k6 workload over the full approved, privacy-cleared Address Quality tagged corpus when available, or a separately approved benchmark-safe copy; preserve stable dataset identity/hash/count and scenario mix | The tracked accuracy CSV and 106-row historical result set are not automatically an approved public load corpus. Do not print raw addresses. If reusing benchmark cases, replay them at controlled, explicit rate/concurrency and avoid accidentally treating sequential accuracy-run results as load evidence. |
| Load | Existing five-address k6 script as an initial repeatable smoke/load sanity case; later replace/extend with approved bounded corpus sampling and declared scenario weights | Five synthetic examples are hand-authored checks, not production traffic. No repo evidence establishes real traffic proportions, request-length distribution, source mix, or concurrency profile. Do not claim representativeness. |
| Stress | Explicitly opt-in, isolated non-production target only, with hard maximum VUs/rate/duration and a separate confirmation guard | Never auto-target production. No stress workload or capacity threshold is currently defined; stress result is exploratory and not a pass/fail SLO. |

The baseline workload must state whether cache state is cold or warm. Since V1 lazily loads large location indexes once, a run that includes the first request measures startup initialization mixed into client latency; a warmed steady-state run answers a different question. Report both only when the experiment intentionally captures both, and never compare them as equivalent.

## Existing datasets and privacy limits

- `tests/api/cases/example.csv` is a four-row sample fixture and is suitable only for script smoke/check wiring.
- `tests/api/cases/address-tagged.csv` has 61 local rows at audit time and is ignored by Git. It carries ground truth/tagging and may contain private inputs. It can be evaluated for an approved baseline workload locally, but must not be copied into a committed public artifact or raw-address report.
- `tests/api/cases/address.csv` has 109 local rows at audit time, is ignored, and README identifies it as a private full dataset. It is not available to reviewers from the repo and must not be called production-representative.
- The committed accuracy JSON contains 106 raw addresses/quality outputs per file; these are historical accuracy artifacts, not workload evidence or a current performance baseline.
- The current k6 load fixture has five hand-selected strings. It is deterministic but too small and homogeneous to support broad traffic-representativeness claims.

## Existing performance documentation

`AGENTS.md` defines ownership (k6 client metrics), missing-data handling, artifact needs, and task order. `tests/api/README.md` documents k6 prerequisites, configuration, smoke/load commands, current load scenario, JSON summary, and current thresholds. `tests/api/page/README.md` documents the accuracy page, its optional k6 section, metadata, and private `full-benchmark.html`. There is no prior dedicated performance audit/contract in `docs/performance/`.
