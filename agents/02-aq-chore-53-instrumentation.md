# aq-chore-53 — performance instrumentation (handoff)

Status: **DONE** (implementation + verification). Follows `agents/01-performance-contract.md` section B and C.

## Task status

| # | Task | Status |
|---|------|--------|
| 1 | Repository/performance audit | DONE (`agents/00-current-state.md`) |
| 2 | Performance measurement contract | DONE (`agents/01-performance-contract.md`) |
| 3 | aq-chore-53 instrumentation | **DONE (this document)** |
| 4 | aq-chore-54 k6 benchmark | TODO — next |
| 5 | Benchmark artifact integration | TODO |
| 6 | aq-chore-52 Performance Report | TODO |
| 7 | End-to-end verification | TODO |
| 8 | Continuous observability stack | Separate future task (out of scope) |

## What changed

New package `internal/metrics`: a concurrency-safe, in-process, cumulative-aggregate collector with a copy-out snapshot. Stdlib only (`sync`, `runtime`, `sort`, `time`), no exporter dependency, no third-party metrics library.

Call sites:

- `internal/logger/logger.go` `EchoMiddleware` — request count, duration, status.
- `internal/service/validate_v1.go` — five stage timers.
- `internal/service/validate_helper.go` — `postal_code_fallback` stage timer.
- `internal/router/router.go` — loopback-only snapshot route.
- `internal/config/config.go` — `ENABLE_PERF_SNAPSHOT` flag, default **false**.

No matching, scoring, candidate generation, evidence, postal, or API behavior was altered. Timers wrap existing statements; no condition, return, or data flow was added or removed. `go test ./internal/service` (17.5s, the validation suite) passes unchanged.

## Snapshot acquisition path

`GET /internal/perf-snapshot` → `metrics.Collect()` JSON.

Access rules (all must pass, else **404**):

1. `ENABLE_PERF_SNAPSHOT=true` — route is not even registered otherwise.
2. `RemoteAddr` is loopback.
3. No `X-Forwarded-For` / `X-Real-IP` header.

Rule 3 exists because `deploy/nginx.example.conf` proxies to the app on localhost: without it every client would look loopback and the snapshot would be public. The harness must talk to the app socket directly. The route is not behind `X-API-Key`, so a local harness needs no secret.

Aggregates are **cumulative**: snapshot before and after the measured phase and difference them. That is how a run isolates its own request/stage work from warmup and earlier traffic.

## Measurements implemented

Snapshot field → contract name:

| Snapshot field | Contract name | Notes |
|---|---|---|
| `http.requests_total` | `app.http_requests_total` | Per (method, route, status) and total |
| `http.duration_ms` | `app.http_request_duration_ms` | count + sum + fixed-bucket histogram |
| `http.errors_4xx_total`, `http.errors_5xx_total` | `app.http_errors_total` | Kept separable |
| `validation_stages.<stage>` | `app.validation_stage_duration_ms` | count + sum + histogram per stage |
| `validation_stages.<stage>.count` | `app.validation_stage_calls_total` | Derived, no second counter |
| `runtime.go_version`, `process_uptime_seconds`, `goroutines`, `heap_alloc_bytes`, `heap_inuse_bytes`, `total_alloc_bytes`, `gc_cycles`, `gc_pause_ns` | `runtime.*` | Point sample / cumulative for deltas |
| `runtime.cpu_count`, `os`, `arch` | `environment.cpu_count`, `environment.os_arch` | Server side; runner captured separately |

`duration_buckets_ms` (1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000 ms + overflow) is published once per snapshot and shared by every aggregate. `count: 0` means *not observed*, never a measured zero; there is deliberately no `avg_ms` field so a zero-observation average cannot be misread.

### Stage boundaries as implemented

| Stage | Wraps |
|---|---|
| `source_and_cache_ready` | `FindSourceByCode` + `ensureEntitiesCachesLoaded` |
| `evidence_resolution` | `ExtractEvidence` + road-context detection + `ResolveEvidence` |
| `candidate_build` | `buildCandidates` (discover, dedupe, enrich, conclusions) |
| `contextual_recovery` | `RecoverContextualEvidence` + its single conditional rebuild |
| `candidate_evaluation` | evaluation loop + ranking sort |
| `postal_code_fallback` | postal lookup **only when invoked** |

`postal_code_fallback` is timed inside `resolveLocationByPostalCode`, after its no-postal-code early return, so the series only exists when the lookup path really ran. Verified live: absent for inputs without a 5-digit postal code, present for inputs with one.

### Deliberately not implemented (contract-optional)

- `app.validation_candidates` — optional work-size observation. Add when a candidate-count question appears; it needs its own bounded numeric histogram.
- `app.contextual_recovery_outcome_total` — optional `recovered|none|error` counter. The stage timer answers "is recovery expensive"; add the outcome split when we need "how often does it change the result".
- `environment.process_rss_bytes`, `cpu_usage`, `memory_limit_bytes`, `container_cpu_limit` — need an OS/container sampler scoped to this process. Per contract these stay NOT MEASURED; do not substitute host memory or load average.

## Cardinality and privacy bounds

Three whitelists, everything else collapses to a fixed bucket:

- **Stage** → known stage enum, else dropped (`ObserveStage` ignores unknown).
- **Method** → GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS, else `OTHER`.
- **Route** → the 5 registered Echo templates, else `other`.

The route whitelist is the privacy control: a raw path or address can never become a label. New routes must be added to `knownRoutes` in `internal/metrics/metrics.go`; a growing `other` series is the signal one was missed. No request ID, IP, host, URI, address, postal code, evidence token, candidate name, or API key reaches the snapshot.

## Verification performed

- `gofmt` clean on all touched files.
- `go build ./...` and `go vet ./...` — clean.
- `go test ./... -count=1` — all packages pass (incl. `internal/service` 17.5s).
- `go test ./internal/metrics -race` — pass.
- Live against local `db/*.db`:
  - `POST /v1/validate` returns 200 with the same response shape (`address_id`, `status`, `confidence`, `formatted_address`, `location`, `assessment`, `resolution`, `metadata`, `normalized_input`, `raw_input`).
  - Snapshot from `127.0.0.1` → 200; from LAN `192.168.100.96` → 404; with `X-Forwarded-For` → 404; with `X-Real-IP` → 404.
  - Flag absent → 404 even with a valid API key (route not registered).
  - Snapshot JSON contains no input address, city, or postal code.
  - A rejected request recorded `errors_4xx_total: 1` with status 404, proving the error-status resolution works.

The metrics self-check (`internal/metrics/metrics_test.go`) is the runnable guard: concurrency, snapshot independence, unknown-stage rejection, bounded method/route, and no label leakage. It already caught one real bug during development — raw route strings were becoming labels before the `knownRoutes` whitelist was added.

## Assumptions

- Server-side duration measured at `logger.EchoMiddleware` is registered first via `e.Use`, so it wraps Recover, CORS, and the route chain. It excludes Echo's `HTTPErrorHandler` write, which happens after the middleware returns.
- Because handler errors are only written after the middleware returns, `res.Status` is not final inside it. Status is resolved from the returned error (`*echo.HTTPError` code, else 500). Without this, every server error would have been recorded as 200.
- `postal_code_fallback` placed inside the helper rather than at the `ValidateAddressV1` call site, purely so it cannot record the no-op path. Documented here because the contract named `ValidateAddressV1` as the location.

## Unresolved issues

1. **Local `vendor/` is stale and untracked.** Plain `go build ./...` fails with "inconsistent vendoring"; verification used `GOFLAGS=-mod=mod`. Not caused by this change and not touched. `go mod vendor` would repair this checkout, or delete `vendor/`. Fresh clones are unaffected.
2. **Pre-existing log quirk:** requests that return a handler error log `status: 0` (or the default 200) and `bytes_out: 0`, because the log runs before Echo writes the error response. The metric is correct; the log line is misleading. Out of scope here, worth a separate fix.
3. **Contract open decision #2 is answered** by the loopback snapshot route, but if the benchmark runner ever runs in a different container from the API, loopback will not reach it and a different transport is needed. Keep the harness and API on the same host/network namespace.
4. **Contract open decisions #1 and #3 are unchanged** — canonical workload corpus and whether the 1% k6 failure threshold is an SLO. Nothing here answers them.

## Recommended next step

**aq-chore-54 (k6 benchmark)**: extend `tests/api/summary.js` / `run-k6.sh` to export bounded status counts and scenario/config metadata, and have the runner take two `metrics.Collect()` snapshots (before/after the measured phase) around the five-address workload. Keep k6 authoritative for client latency; use these measurements only to explain server-side time. No stress script without a non-production target guard and hard caps.
