# aq-chore-53 — performance instrumentation (handoff)

Status: **DONE** (implementation + verification). Follows `agents/01-performance-contract.md` sections B and C.

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
| 8 | Continuous observability stack | Prometheus/Grafana servers still future work |

## Plan change (user-approved)

The first implementation used a small hand-rolled collector (`internal/metrics`). The user directed: use OpenTelemetry to measure time spent per span instead of a new package doing the same, and target Prometheus + Grafana. `internal/metrics` was deleted and replaced by `internal/telemetry`, built on the OTel SDK.

Scope of that approval: **application-side OTel**. No OpenTelemetry Collector is installed (the brief still forbids it), and no Prometheus or Grafana server is installed by this task. The app exports Prometheus exposition format directly, so Prometheus can scrape it and Grafana can chart it whenever those are provisioned.

## What changed

New package `internal/telemetry` (OTel SDK + Prometheus exporter + `contrib/instrumentation/runtime`). Call sites:

- `internal/logger/logger.go` `EchoMiddleware` — one `http.request` span per request.
- `internal/service/validate_v1.go` — five stage spans.
- `internal/service/validate_helper.go` — `stage.postal_code_fallback` span.
- `internal/router/router.go` — `GET /metrics` (Prometheus).
- `internal/config/config.go` — `ENABLE_METRICS`, default **false**.
- `cmd/server/main.go` — `telemetry.Init()` at startup, shutdown at exit.

Deleted: `internal/metrics/` (collector and its test). Tests were ported to `internal/telemetry/telemetry_test.go`.

No matching, scoring, candidate generation, evidence, postal, or API behavior was altered. `go test ./internal/service` (17.6s, the validation suite) passes unchanged.

## Timing model: spans are the source of truth

The SDK stamps start and end on each span. This package never calls `time.Now` to time work; it reads the stamp pair off the completed span:

```
validate stages + http.request  ->  trace spans (SDK timestamps)
                                        |
                          spanMetrics processor (OnEnd)
                                        v
                    app.span.duration_ms  histogram
                                        |
                          otel Prometheus exporter
                                        v
                    GET /metrics  ->  Prometheus  ->  Grafana
```

Consequence for aq-chore-54: a request contributes exactly one duration sample per span, and there is no second timing path to disagree with it. k6 remains authoritative for client-observed latency.

## Measurements

Exported metric: `app_span_duration_ms` (histogram, OTel default buckets). Labels: `span.name`, `span.status`, `http.status_class`.

| Contract name | Source |
|---|---|
| `app.http_request_duration_ms` | `app_span_duration_ms{span_name="http.request"}` |
| `app.http_requests_total` | `_count` of the same series |
| `app.http_errors_total` | `_count{http.status_class="4xx"}` and `"5xx"` — separable |
| `app.validation_stage_duration_ms` | `app_span_duration_ms{span.name="stage.*"}` |
| `app.validation_stage_calls_total` | `_count` per stage |
| `runtime.goroutines` | `go_goroutine_count` |
| `runtime.heap_alloc_bytes` / `total_alloc_bytes` | `go_memory_used`, `go_memory_allocated_total` |
| `runtime.gc_cycles`, `runtime.gc_pause_ns` | **not_measured** — see below |
| `environment.cpu_count` | `go_processor_limit` |
| `runtime.go_version`, `os`, `arch` | **not_measured** from `/metrics` — see below |

### Not measured, deliberately

- **`gc_cycles`, `gc_pause_ns`** — `go.opentelemetry.io/contrib/instrumentation/runtime` v0.71.0 exposes only `WithMeterProvider` and `WithMinimumReadMemStatsInterval`; it emits no GC pause or cycle instrument. Emitting them would mean hand-scraping `runtime.MemStats`, which is exactly the hand-rolled path the user ruled out. Mark `not_measured` in the artifact.
- **`runtime.go_version`, `environment.os_arch`** — not in the Prometheus runtime set. Available from the Go toolchain at build time; aq-chore-54 can stamp them into the runner metadata instead of scraping them.

### Stage boundaries as implemented

| Span | Wraps |
|---|---|
| `stage.source_and_cache_ready` | `FindSourceByCode` + `ensureEntitiesCachesLoaded` |
| `stage.evidence_resolution` | `ExtractEvidence` + road-context detection + `ResolveEvidence` |
| `stage.candidate_build` | `buildCandidates` (discover, dedupe, enrich, conclusions) |
| `stage.contextual_recovery` | `RecoverContextualEvidence` + its single conditional rebuild |
| `stage.candidate_evaluation` | evaluation loop + ranking sort |
| `stage.postal_code_fallback` | postal lookup **only when invoked** |

`stage.postal_code_fallback` starts inside `resolveLocationByPostalCode`, after its no-postal-code early return, so the series only exists when the lookup path really ran. Verified live: present for inputs carrying a 5-digit postal code, absent otherwise.

`span.status` is `error` only for 5xx handler errors. Stage spans stay `unset`, because a stage returning no candidate is not an error.

## Cardinality and privacy bounds

Three whitelists in `internal/telemetry`, everything else collapses to a fixed bucket:

- **Span name** → the 7 fixed spans, else `other`.
- **Method** → GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS, else `OTHER`.
- **Route** → the 5 registered Echo templates, else `other`.

`http.status_class` is 5 fixed values (`2xx`/`3xx`/`4xx`/`5xx`/`other`, plus `none` for non-HTTP spans) and rejects non-integer attributes. The processor emits only these four labels — a span's own attributes are never copied to metrics, so no caller can smuggle an address into a label. New routes must be added to `knownRoutes`; a growing `other` series is the signal one was missed.

## Acquisition path

`GET /metrics` (Prometheus exposition format), registered only when `ENABLE_METRICS=true`. Default false, so nothing diagnostic is exposed unless asked for.

This replaced `/internal/perf-snapshot` (loopback-only JSON). Unlike it, `/metrics` has **no loopback restriction** — Prometheus scrapes remotely, which is the point. Deployment must network-restrict it. Labels are bounded, so even exposed it carries no addresses or request data.

## Verification performed

- `gofmt` clean on all touched files. `go build ./...` and `go vet ./...` clean.
- `go test ./... -count=1` — all packages pass (incl. `internal/service` 17.6s).
- `go test ./internal/telemetry -race` — 5 tests pass: span becomes a Prometheus metric, bounded dimensions, status-class mapping, no label leakage, concurrent spans.
- Live against local `db/*.db`:
  - `POST /v1/validate` returns 200, same response shape as before.
  - `/metrics` 200 and shows all 7 span names.
  - `http_status_class` separates 2xx and 4xx on `span_name="http.request"`.
  - Zero matches for the test address, city, or postal code in the exposition output.
  - Runtime gauges present (`go_goroutine_count`, `go_memory_used`, `go_config_gogc`, `go_processor_limit`).

The self-check caught two real defects during development: raw route strings becoming labels (fixed with the `knownRoutes` whitelist), and `telemetry.Init()` never being wired into `main.go` (found by the live run, not the unit tests — the tests call `Init()` themselves).

## Assumptions

- Request duration is measured in `logger.EchoMiddleware`, registered first via `e.Use`, so it wraps Recover, CORS and the route chain. It excludes Echo's `HTTPErrorHandler` write, which happens after the middleware returns.
- Handler errors are written after the middleware returns, so `res.Status` is not final inside it. Status is resolved from the returned error (`*echo.HTTPError` code, else 500); without this every server error would be recorded as 2xx. Verified live: a 401 lands in `http_status_class="4xx"`.
- `Start` deliberately discards the child context, so no call site changes behaviour. Nothing in the request path consumes trace context yet.

## Unresolved issues

1. **Local `vendor/` is stale and untracked.** Plain `go build ./...` fails with "inconsistent vendoring"; verification used `GOFLAGS=-mod=mod`. Not caused by this change. `go mod vendor` or deleting `vendor/` repairs the checkout.
2. **Pre-existing log quirk:** requests returning a handler error log `status: 0` and `bytes_out: 0`, because the log runs before Echo writes the error response. The metric is correct; the log line is misleading. Out of scope here.
3. **Contract open decisions #1 and #3 unchanged** — canonical workload corpus, and whether the 1% k6 failure threshold is an SLO.
4. **`prometheus/client_golang` advisory noise** — the scanner flags GHSA-cg3q-j54f-5p7p and GO-2022-0322 against this module. GO-2022-0322 affects pre-1.11.1; the resolved version is current. Worth a re-check before release.

## Recommended next step

**aq-chore-54 (k6 benchmark)**: run k6 over the five `synthetic-five-address` items and assemble the artifact from `/metrics`. Take two scrapes (before/after the measured phase) and difference the counters and histogram sums, which is how a run isolates its own work from warmup and cold start. Record `gc_cycles`, `gc_pause_ns`, `runtime.go_version`, `environment.os_arch` as `not_measured` rather than substituting anything. Keep k6 authoritative for client latency.
