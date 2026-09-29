# Address Quality performance measurement contract

Status: Stage 1 recommendation for aq-chore-53, aq-chore-54, benchmark-artifact assembly, and aq-chore-52. This contract records no performance target or measured result. The contract must not change V1 matching, scoring, validation behavior, or the public API.

## Purpose and ownership

This is a contract for controlled benchmark experiments: **what happened during this run?** It is not continuous operational monitoring.

- **k6 owns client-observed HTTP performance**: request count, response latency distribution, throughput, failures/statuses, and configured/observed virtual users. Its raw JSON summary is the source for client metrics.
- **Address Quality application instrumentation owns server-side diagnostic context**: bounded request outcomes and a small set of coarse validation-stage durations that explain the server's work. It does not replace k6 latency or calculate client percentiles.
- **Runtime/environment snapshots describe only captured facts**. If no reliable sampler exists or the collection mechanism was not enabled, represent the value as unavailable (`null` plus a reason or a `not_measured` record), never zero.
- **A versioned JSON benchmark artifact** is assembled from k6's machine-readable summary, application snapshot, runtime/environment capture, and workload/run metadata. A report consumes this artifact, never console output.
- Prometheus, Grafana, and an OpenTelemetry Collector are out of scope. Do not install/deploy/configure them in these tasks.

Existing references: `AGENTS.md`; `tests/api/load-test.js`, `tests/api/summary.js`, `tests/api/run-k6.sh`; `tests/api/benchmark-test.js`; `tests/api/page/build.js`; `internal/logger/logger.go`; `internal/service/validate_v1.go`.

## A. k6 / client-observed measurements

The current k6 summary already captures several built-in metrics and threshold outcomes, but the artifact assembler must validate their presence and make the contract explicit. k6 output is the authority; do not estimate client metrics from Echo logs or accuracy-benchmark output.

| Canonical name | Definition / unit | Source and collection | Dimensions/cardinality | Baseline report | Limitations |
|---|---|---|---|---|---|
| `k6.http_requests` | Count of HTTP requests sent; requests | k6 built-in `http_reqs`, exported in `handleSummary` JSON | Scenario and method/path if explicitly tagged; avoid raw URL/query | REQUIRED | Count includes selected endpoints only; document included checks and workload. |
| `k6.http_request_duration_ms` | Request duration trend and its p50, p90, p95, p99; milliseconds | k6 built-in `http_req_duration`, `summaryTrendStats`; explicitly expose p50 (existing `med`) as `p(50)` or unambiguous `p50` | Fixed scenario/endpoint tags only | REQUIRED | k6 duration excludes time outside HTTP request lifecycle; aggregate percentile across unlike scenarios can mislead. Preserve per-scenario groups where scenarios differ. |
| `k6.http_requests_per_second` | HTTP request count divided by the run's measured elapsed seconds; requests/s | k6 built-in `http_reqs` rate/value from the structured summary; derive from count and exact measured run duration if the exported rate is absent, and document the formula | Same dimensions as request count | REQUIRED | This whole-run mean includes ramp-up/down for the existing staged test. Do not present it as steady-state throughput; show phase/scenario rates only when collected explicitly. |
| `k6.http_failed_rate` | Fraction of HTTP requests k6 classifies as failed; ratio [0,1] | k6 built-in `http_req_failed` | Scenario only | REQUIRED | Keep separate from functional check failure and status-code counts. k6's classification is not synonymous with non-2xx. |
| `k6.http_status_count` | Count of responses by HTTP status; responses | Bounded k6 `Counter` custom metric with a status tag, or a deterministic summary reduction preserving status counts | HTTP status only (small fixed code set) and optional scenario | REQUIRED | Do not emit address, request ID, API key, arbitrary error, or high-cardinality path labels. Verify behavior for connection errors with no response code; those remain covered by failed rate. |
| `k6.check_pass_rate` | Passed checks / total checks; ratio [0,1] | k6 built-in `checks` and flattened `root_group` in summary | Fixed check name/group; do not create per-address checks | REQUIRED | Functional correctness signal for test assertions, not an HTTP reliability SLO. |
| `k6.vus_configured` | Target virtual users per stage; VUs | Script `options` retained in summary | Scenario/stage | REQUIRED | Describes planned load, not achieved simultaneous active VUs. |
| `k6.vus_observed_max` | Maximum observed active VUs during run; VUs | k6 built-in `vus_max` or `vus` summary, whichever is present and semantically correct | Scenario | OPTIONAL | Include only when present and correctly defined; do not synthesize it from configuration. |
| `k6.run_duration_ms` | Actual elapsed k6 run duration; milliseconds | `data.state.testRunDurationMs` in `summary.js` | Whole run | REQUIRED | Not interchangeable with stage sum unless verified; this is the denominator for derived throughput. |

Do not hard-code pass/fail targets in this contract. The existing load script has `http_req_failed < 1%` and `checks == 100%`, but these are current script thresholds, not approved service SLOs. Preserve/report them as *configured experiment thresholds* until a human-approved task defines a different target; do not claim they are service objectives.

## B. Application measurements

Application data is server-side diagnostic evidence. Implement it with a concurrency-safe, low-overhead in-process collector/snapshot that has no required external exporter and is safe if later adapted to Prometheus or OpenTelemetry. Do not add a publicly exposed diagnostic endpoint by default. The benchmark harness may read a benchmark-mode/local-only snapshot or use another explicitly access-restricted mechanism; its exact transport is implementation work and must not leak to public callers. If the runtime/app snapshot cannot be acquired, record it unavailable and continue to preserve the k6 result.

Metric names below are contract names, not a demand to add a third-party metrics dependency. Use fixed dimensions; keep application and k6 values distinct even when their names describe analogous concepts.

| Canonical name | Definition / unit | Source and expected collection | Dimensions/cardinality | Baseline report | Limitations |
|---|---|---|---|---|---|
| `app.http_requests_total` | Number of requests entering the Echo request logging/measurement boundary; requests | Increment once in an app-owned middleware; exclude or explicitly include health and distinguish it with a bounded route class | Method, route template, status class/status code only; no raw URI, IP, host, request ID, API key | REQUIRED | Server count may differ from k6 count due to network failures, rejected auth/rate limit, health probes, or measurement window. Declare included route classes. |
| `app.http_request_duration_ms` | Server-side duration of the Echo chain; milliseconds | Monotonic timer in middleware; aggregate count/sum and fixed-bucket histogram or snapshot sufficient to report distribution if implemented | Method, route template, status class; fixed buckets | REQUIRED as diagnostic aggregate; never client-performance authority | Current `EchoMiddleware` logs a per-request duration but has no aggregation. Server boundary differs from k6 and includes/excludes middleware according to placement; document placement. |
| `app.http_errors_total` | Server requests producing 4xx/5xx or handler errors; requests | Derive from the same measured request outcome/status. Keep 4xx and 5xx separable | Route template and bounded status/status class | REQUIRED | Do not equate validation 400s with unexpected server failures. Preserve full status distribution in k6; app-side error counter is only a diagnostic cross-check. |
| `app.validation_stage_duration_ms` | Duration per accepted coarse V1 stage; milliseconds | Timers around the bounded stage boundaries below, aggregated as count/sum/histogram per stage | Fixed `stage` enum; optionally outcome `ok/error`; no source/address/token/candidate ID labels | REQUIRED to answer where service-side regression occurs; stage-specific report may show unavailable until collected | Instrumentation changes must not alter result/control flow. Aggregate metrics do not provide per-request spans or guarantee percentile accuracy unless buckets/snapshot provide it. |
| `app.validation_stage_calls_total` | Number of completed stage invocations; calls | Increment alongside stage duration | Fixed stage enum and bounded outcome only | OPTIONAL | Useful for conditional-stage interpretation; do not count sub-function calls as stage calls. |
| `app.validation_candidates` | Number of candidate records at defined candidate-build/scoring boundary; candidates/request observation | Numeric distribution/summary captured at candidate-build or evaluation stage, not label | No labels other than fixed stage | OPTIONAL | Useful to explain work-size changes; a per-request metric is sampled/aggregated, not a stable workload distribution without an explicitly defined histogram. |
| `app.contextual_recovery_outcome_total` | Count of requests where second-pass recovery recovered evidence vs none; requests | Increment fixed outcome in the contextual recovery stage | `outcome={recovered,none,error}` only | OPTIONAL | Do not expose token spans or correction strings; timer alone is the REQUIRED diagnostic measure. |

### Stable V1 stage boundaries

Use a minimal stable set in `ValidateAddressV1`, not timers around every helper:

1. `source_and_cache_ready`: source lookup plus one-time cache readiness. Report separately from warmed request stages only if cold-start measurement is intentionally captured; `sync.Once` means the first request differs materially.
2. `evidence_resolution`: evidence extraction, road-context handling, and exact/compact entity resolution. Combined stage answers whether parsing/resolution changed without over-instrumenting internals.
3. `candidate_build`: discovery, deduplication, hierarchy enrichment, and conclusions. Candidate count is an optional numeric observation.
4. `contextual_recovery`: second-pass in-memory neighborhood recovery, including conditional single rebuild. Outcome can distinguish recovery-active from no-op without exposing input values.
5. `candidate_evaluation`: complete candidate evaluation plus ordering/ranking, one stage around the loop and sort (not per candidate).
6. `postal_code_fallback`: conditional post-winner lookup, instrument only when that lookup path is actually invoked; no postal code label. Optional because it is not present in every request.

Sanitize/normalize is not a separate baseline metric: its standalone diagnostic value is unproven and it is likely too fine to justify another series before measurements identify it as a bottleneck. Async request storage is outside the synchronous validation-response latency boundary; do not include its DB write latency in a response stage. If request persistence is explicitly enabled in an experiment, report that setting and optionally capture bounded queue rejection/failure counts in a later contract revision.

Do not add a V1 external dependency latency metric: no external lookup exists in the inspected V1 path. Google Maps calls/cache belong to V0 and are outside this baseline.

## C. Runtime and environment measurements

Collect stable facts in the benchmark runner/artifact where available, and take a Go runtime snapshot at a documented time (prefer immediately before and after the measured phase, labeling each). Go's `runtime` package can provide runtime version, goroutine count, heap allocation/in-use, allocation totals, GC count/pause totals, and uptime from a process-start timestamp. These are snapshots/deltas, not time-series profiles. They are optional diagnostics and must never be inferred from k6.

| Canonical name | Definition / unit | Source/mechanism | Labels/cardinality | Baseline report | Limitations |
|---|---|---|---|---|---|
| `runtime.go_version` | Go toolchain/runtime version string | `runtime.Version()` in app snapshot | None | REQUIRED metadata | Runtime version does not uniquely define compiler flags or binary build. |
| `runtime.process_uptime_seconds` | Seconds since application process start | Captured startup timestamp and current monotonic time | Snapshot phase only | REQUIRED metadata/snapshot | Distinguish app uptime from k6 duration; capture start and end if available. |
| `runtime.goroutines` | Current goroutine count; goroutines | `runtime.NumGoroutine()` | Snapshot phase | OPTIONAL | A point sample, not peak or leak proof. |
| `runtime.heap_alloc_bytes` | Current allocated heap bytes; bytes | `runtime.ReadMemStats` or equivalent runtime metrics | Snapshot phase | OPTIONAL | Point sample; GC timing affects comparison. |
| `runtime.heap_inuse_bytes` | Heap in-use bytes; bytes | Runtime memory stats | Snapshot phase | OPTIONAL | Not total process RSS/container memory. |
| `runtime.total_alloc_bytes` | Cumulative bytes allocated since process start; bytes | Runtime memory stats | Before/after delta when both captured | OPTIONAL | Delta depends on exact interval and includes non-request work. |
| `runtime.gc_cycles` | Cumulative completed GC cycles; cycles | Runtime memory stats/runtime metrics before/after delta | Before/after delta | OPTIONAL | A delta, not pause distribution or causal attribution. |
| `runtime.gc_pause_ns` | Cumulative GC pause time; nanoseconds | Runtime memory stats/runtime metrics before/after delta | Before/after delta | OPTIONAL | Aggregate interval only; no per-request pause attribution. |
| `environment.process_rss_bytes` | Process resident set size; bytes | External OS/container sampler scoped to the API process/container, if available | Snapshot phase | OPTIONAL; otherwise NOT MEASURED | Host-level memory is not process memory. Must identify sampler and target process/container. |
| `environment.cpu_usage` | CPU time or utilization for the API process/container over the run | OS/container sampler scoped to API process/container, with interval and units | Snapshot interval | OPTIONAL; otherwise NOT MEASURED | A single host CPU percentage is ambiguous; do not use host load averages as process CPU. |
| `environment.cpu_count` | Logical CPUs visible to the benchmark process/container; count | `runtime.NumCPU()`/container-visible fact, if that is the intended scope | None | REQUIRED metadata | Must distinguish runner CPU from server CPU. |
| `environment.memory_limit_bytes` | Effective process/container memory limit; bytes | Container runtime/cgroup if reliably discoverable; otherwise NOT MEASURED | None | OPTIONAL metadata | Do not substitute host MemTotal. |
| `environment.os_arch` | OS and architecture; strings | Runtime/OS metadata | None | REQUIRED metadata | Record server and k6 runner separately. |
| `environment.container_cpu_limit` | Configured/effective CPU quota; CPU units | Deployment config or container metadata if available | None | REQUIRED if known; otherwise NOT MEASURED | `deploy/docker-compose.prod.yml` specifies 1 CPU and 512 MiB for that deployment, not necessarily the benchmark target. Record observed target facts, not repo defaults as actual conditions. |

Required runtime metadata means record a value or explicit `not_measured` reason. Resource values must remain absent/null/unavailable when not captured. Do not fill missing with numeric zero.

## Metric explicitly rejected / not implemented for this scope

- Application-side replacement percentiles for k6 client latency; the measurements cross different boundaries.
- Request IDs, raw/normalized addresses, postal codes, evidence tokens, candidate identifiers/names, API keys, client IPs, raw URI, host, arbitrary source codes or error strings as metric labels.
- Per-function/per-token/per-candidate timers, request-by-request traces, CPU profiles, heap profiles, or continuous scraping infrastructure.
- V1 external dependency timing.
- A public `/metrics`, pprof, or debug endpoint without a separate security design and explicit approval. If later required, choose a benchmark-mode or access-restricted mechanism.
- SLOs, service capacity claims, pass/fail thresholds, or production traffic characterization not established by repository evidence or explicit human decision.
- Fake runtime/resource values (including converting unavailable to zero) and raw k6 console scraping.

## Benchmark pipeline and artifact contract

Smallest recommended pipeline:

```text
k6 scripted scenario -> Address Quality API
                         |-- bounded application stage/request snapshot
                         |-- Go/runtime snapshot (if enabled)
                         v
                 k6 summary JSON + run/environment metadata
                         v
                versioned benchmark result JSON
                         v
               static Performance Report consumer
```

The existing `tests/api/summary.js` is the natural k6 export boundary; `tests/api/page/build.js` already consumes structured input. Keep the existing accuracy benchmark artifacts distinct from performance results: `benchmark-test.js` currently emits correctness records, not k6 load evidence. Do not make the report parse human-readable k6 console output. During transition, a new artifact assembler may consume `summary.js`'s JSON and an application snapshot, then write a separate versioned result artifact; do not force it into the accuracy row schema.

Recommend one versioned JSON document per performance run, written atomically after collection and containing at least:

```json
{
  "schema_version": "1.0",
  "metadata": {
    "run_id": "...",
    "started_at": "...",
    "completed_at": "...",
    "git_revision": "...",
    "application_version": "...",
    "scenario": "..."
  },
  "environment": {
    "runner": {},
    "server": {}
  },
  "workload": {
    "target_class": "local|staging|other",
    "api_version": "v1",
    "dataset_id": "...",
    "dataset_sha256": "...",
    "dataset_count": 0,
    "cache_state": "cold|warm|unknown",
    "configuration": {}
  },
  "k6": {},
  "application": {},
  "runtime": {},
  "missing_measurements": [],
  "thresholds": {},
  "observations": []
}
```

This is a conceptual field layout, not a requirement to use those exact nested names beyond `schema_version` and the measurement-domain separation. `dataset_count: 0` above is a schema-shape example only, not a default: actual counts must be measured; unavailable counts are null/unavailable. Prefer `null` plus an explicit `missing_measurements` entry `{name, reason}` for optional scalar fields, never zero. Thresholds are only the values configured for that run, with provenance; an empty/unconfigured threshold set is not a passed threshold. Observations are evidence/context, not pass/fail objectives.

Required artifact validity:

- schema version must be present and supported by the consumer;
- run metadata must identify build/revision, scenario, target, run times, duration, and workload/dataset identity or explain why unavailable;
- client k6 summary must be parsed as structured JSON and checked for required metrics before report publication;
- `missing_measurements` must list every unavailable REQUIRED field and any optional field the run/report elects to show;
- artifact must preserve raw k6 source filename/hash and application snapshot provenance; assembly must fail clearly rather than silently fabricate fields;
- generated result must not embed private raw addresses or API keys;
- write to a temporary file then rename, so interrupted assembly does not leave a valid-looking partial artifact.

## Workload and comparison contract

- Smoke uses the existing one-iteration script only for route/auth/response checks, never as a stable latency sample.
- Initial controlled baseline uses the existing five-address k6 scenario, explicitly labeled `synthetic-five-address`. This is the intentionally scoped performance workload for this task, not production traffic. The accuracy benchmark's private/non-committed dataset and its historical 106-record result are separate and must not be substituted for these five performance inputs.
- Load testing declares stage schedule, VUs, request mix, dataset identity/count, rate/think time, API version, cache state, and whether health probes share the target.
- Stress testing is opt-in only and must require an explicit non-production base URL/target classification. Add a hard upper bound for VUs, rate, and duration; refuse production hosts rather than relying on a developer remembering a warning. No stress workload runs automatically.
- k6 concurrency means virtual users; achieved RPS is a consequence of request duration and any sleep/pacing. Capture both configured stages and observed metrics when present.
- Separate cold-start/cache initialization from warmed steady-state results. Do not compare a run with first-hit cache/database loading to a fully warmed run as if equivalent.

## Performance report input and requirements (aq-chore-52; not built here)

The report consumes the versioned JSON result artifact and renders:

- Executive summary with test identity, validation/schema state, and configured-threshold outcomes (no invented SLO badge).
- Environment: runner and server OS/architecture, Go/k6 version, CPU/memory/container limits when actually captured, deployment/application revision.
- Workload/test configuration: scenario, target class/base URL with secrets stripped, API version, dataset ID/hash/count, duration, VUs/stages, pacing, cache state, and thresholds as configured.
- Latency: k6 client duration distribution and p50/p90/p95/p99 with units, plus optional separate app request/stage diagnostics clearly labeled server-side.
- Throughput and request count from k6, showing measurement interval and scenario grouping.
- Reliability: HTTP failed rate, status counts, check pass rate, and any collection errors separately.
- Resource/runtime: captured process CPU/memory and Go/GC snapshot/deltas; conspicuous NOT MEASURED values for omissions.
- Internal diagnostics: only contracted stage metrics, counts/outcomes, and caveats.
- Before/after comparisons: show deltas only for compatible run pairs; otherwise show “not directly comparable” and explain the differing properties.
- Methodology/audit information: schema version, source artifacts, run timestamp, revision, assembly version, missing fields, data privacy statement, and reproduction command/config reference.

### Comparison gate

Only present a direct numeric before/after comparison when the relevant properties match or a reviewer explicitly marks the difference as intentional and interprets it:

- API version, request path/scenario semantics, application build/revision context, and schema/metric definition;
- same dataset ID and hash, row count, selection order/weights, payload shape, and check behavior;
- same target class and server/container CPU/memory constraints, runtime version/configuration, DB engine and data/source version, and material network path/region;
- same k6 version, runner resource limits, VU/stage schedule, pacing, duration, thresholds, and warmup policy;
- same cache state (cold/warm), storage setting, Postgres enabled/disabled, and external mocks/dependencies;
- same aggregation definitions, sample validity, and error accounting.

If a required comparison dimension is missing, mark comparability `unknown` and suppress a definitive delta. Document any material mismatch; do not normalize away a changed workload or environmental constraint.

## Future observability boundary

Benchmark work owns controlled-run metadata, per-run k6/client measurements, short-lived application/runtime snapshots, reproducible artifacts, and static reports. Continuous observability owns long-term service behavior, alerting, fleet/resource history, and dashboards. Reusable application metric names, units, bounded labels, and histogram boundaries should be transport-neutral so a future Prometheus adapter or OpenTelemetry instruments can export the same signals. Do not add exporter packages or deploy Prometheus, Grafana, or an OpenTelemetry Collector in this phase.

## Implementation handoff and open decisions

Recommended order for MiMo 2.6 Pro:

1. **aq-chore-53, instrumentation**: add concurrency-safe in-process collector and bounded snapshot access; implement only the accepted request/stage measures; add tests for concurrent increments, stage outcomes, no label leakage, and unchanged response behavior. Define a non-public benchmark snapshot acquisition path before coding and verify deployment exposure; no public debug endpoint by default.
2. **aq-chore-54, k6**: keep k6 authoritative; enhance smoke/load scripts only to export the specified status counts and explicit scenario/config metadata, without changing validation payload semantics. Add an opt-in stress script only with non-production target guard and hard caps.
3. **Artifact assembly**: parse k6 summary JSON plus app/runtime/environment snapshots into versioned JSON; validate required fields; list missing values; redact secrets; atomically write. Test valid, missing, malformed, and incompatible schema inputs.
4. **aq-chore-52, report**: add a static consumer of the versioned artifact, not console parsing; display unavailable values honestly and suppress incomparable deltas. Keep it separate from accuracy metrics or clearly separate the report sections/data contracts.
5. **End-to-end verification**: build/test Go; run smoke against local/staging; run a bounded controlled k6 baseline only against an explicitly selected non-production target; verify artifact schema/content and report; check privacy and Git diff; report actual environment and limitations. Do not run stress against production.

The only genuine human decisions not settled by current repository evidence are: (1) which privacy-cleared workload corpus and scenario mix to approve as the canonical baseline (existing test CSVs are private and the five-address fixture is not representative); (2) which non-public mechanism should expose the app snapshot to the benchmark runner, given current API/deployment routing; and (3) whether the existing 1% k6 HTTP failure threshold remains only a test threshold or should later be revised/approved as a target. Do not invent answers or block a local pipeline-only prototype on a false claim of production representativeness.
