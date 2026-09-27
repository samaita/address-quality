# aq-chore-54 — k6 performance benchmark

Status: complete. Machine-readable output is `agents/aq-chore-54-performance-benchmark.json`.
This document is the human analysis. Where the two disagree the JSON is the artifact and this
text is the reading of it.

## Task status

| # | Task | Status | Artifact |
|---|---|---|---|
| 1 | Repository/performance audit | done | `agents/00-current-state.md` |
| 2 | Performance measurement contract | done | `agents/01-performance-contract.md` |
| 3 | aq-chore-53 instrumentation | done | `agents/02-aq-chore-53-instrumentation.md` |
| 4 | aq-chore-54 k6 benchmark | **done** | `agents/aq-chore-54-performance-benchmark.json` + this file |
| 5 | Benchmark artifact integration | next | — |
| 6 | aq-chore-52 Performance Report | pending | — |
| 7 | End-to-end verification | pending | — |
| 8 | Observability stack (Prometheus + Grafana) | future, separate task | exporter is in place, servers not installed |

## What ran

| Field | Value |
|---|---|
| Script | `tests/api/load-test.js` |
| Runner | k6 v2.2.0 |
| Target | `http://127.0.0.1:7394`, local dev loopback, `ENABLE_METRICS=true`, not production |
| Workload | `synthetic-five-address`, 5 addresses, round-robin |
| Scenario | staged ramp 10s to 10 VU, 30s hold, 10s to 0 VU |
| Observed duration | 50.02 s |
| Requests | 2814 iterations, 2814 HTTP requests, 56.26 req/s |
| Revision | `fa1dd688b50b20e2e7fa4104d5dd72a6eb5789f9` (`feat/performance`) |
| Run window | 2026-09-27T04:47:17Z to 2026-09-27T04:48:10Z (11:47 to 11:48 WIB) |

The workload file was extracted first (`fa1dd68`) so k6 and the artifact share one source for
dataset identity and count.

## Client-observed results (k6 is authoritative)

All thresholds passed: `http_req_failed` 0.00 percent, `checks` 100 percent (5628 of 5628).
Zero HTTP errors, zero non-200 responses.

| Metric (ms) | avg | min | med | max | p90 | p95 | p99 |
|---|---|---|---|---|---|---|---|
| `http_req_duration` | 141.47 | 4.36 | 157.34 | 526.79 | 240.18 | 269.81 | 363.23 |
| `http_req_waiting` | 139.78 | 3.88 | 155.58 | 526.00 | 236.30 | 268.41 | 362.35 |
| `iteration_duration` | 144.06 | 5.46 | 159.92 | 529.85 | 242.45 | 273.15 | 365.81 |

Almost all of `http_req_duration` is `http_req_waiting`, so the time is server work, not transfer.

These percentiles aggregate ramp-up, hold and ramp-down. They are not a pure steady-state slice.

## Cold start is a separate world

Measured as the first `POST /v1/validate` on a freshly started process, before any warm-up.

**Cold request: 5.428 s.** Warm median: 157 ms. That is a 35x gap.

| Cold stage | ms | share |
|---|---|---|
| `stage.source_and_cache_ready` | 4400.12 | 81.15% |
| `stage.evidence_resolution` | 933.80 | 17.22% |
| `stage.contextual_recovery` | 87.24 | 1.61% |
| `stage.candidate_build` | 1.07 | 0.02% |
| `stage.candidate_evaluation` | 0.23 | 0.00% |
| `stage.postal_code_fallback` | 0.004 | 0.00% |

The lazy cache load is the cold cost: 81 percent of the request sits in
`stage.source_and_cache_ready`. `stage.contextual_recovery` is also 111x its warm value (87 ms vs
0.79 ms) because the fuzzy neighborhood builds on first use.

Cold numbers are one sample. They are a flag, not a distribution.

## Server-side stages, warm steady state

Diagnostic only. k6 owns the client view above. These are span-duration sums from
`app_span_duration_ms`, computed as the delta between two `/metrics` scrapes around the run.

| Stage | count | sum (ms) | per call (ms) | share of stage time |
|---|---|---|---|---|
| `stage.evidence_resolution` | 2814 | 324220.26 | 115.22 | **94.64%** |
| `stage.source_and_cache_ready` | 2814 | 8641.50 | 3.07 | 2.52% |
| `stage.candidate_build` | 2814 | 5800.31 | 2.06 | 1.69% |
| `stage.contextual_recovery` | 2814 | 2215.69 | 0.79 | 0.65% |
| `stage.candidate_evaluation` | 2814 | 1676.86 | 0.60 | 0.49% |
| `stage.postal_code_fallback` | 2250 | 33.08 | 0.015 | 0.01% |

Per-call figures are derived here for readability and are deliberately absent from the JSON
artifact: a zero-observation average must not read as a measured zero.

Stages explain 97.31 percent of `http.request` time (125.07 ms per call over 2815 spans). The
missing 2.7 percent is sanitization, normalization, JSON encoding and HTTP server overhead that
no stage wraps.

### What the stage data says

1. **`stage.evidence_resolution` is 94.64 percent of server-side stage time.** It wraps
   `ExtractEvidence` + `ResolveEvidence`. This is the first evidence-driven pointer at where
   request time actually goes. It is a measurement, not a change: the contract forbids altering
   evidence semantics without explicit approval, and nothing here justifies doing so yet.
2. **The cold path is dominated by cache loading**, not by validation work. If cold latency ever
   matters, pre-loading the caches at startup would address 81 percent of it. Not done: that is a
   startup behavior change and needs a decision.
3. **The rebuild path is unexercised by this workload.** `stage.candidate_build` has
   `calls_per_request` exactly 1.000 across all 2814 requests, so contextual recovery never added
   evidence and the second build never ran. The two-samples-per-request shape is real code but has
   no measurement from this workload. Do not report its cost from this run.
4. **`stage.postal_code_fallback` fires on 0.80 of requests** (2250 of 2814), matching 4 of the 5
   addresses carrying a postal code. It costs 0.015 ms per call and is not a factor.

## Runtime characteristics

From the post-load `/metrics` scrape, so a point-in-time gauge, not a time series.

| Measurement | Value |
|---|---|
| `GOGC` | 100 |
| `GOMAXPROCS` (`go_processor_limit`) | 4 |
| Goroutines | 23 |
| Heap/other used | 194.14 MB |
| Stack used | 1.18 MB |
| GC goal | 205.87 MB |
| Total allocated (cumulative) | 472.20 MB |
| Allocation count (cumulative) | 5684438 |

`gc_cycles`, `gc_pause_ns`, `go_version` and `os_arch` are `not_measured`. The OTel contrib runtime
instrumentation exposes no GC instruments and there is a deliberate decision not to hand-scrape
`runtime.MemStats`. Never read these as zero.

## Safety work in this task

`tests/api/load-test.js` gained two guards, both required by the measurement contract and both
verified before the run:

- **Production refusal guard.** `config.js` resolves `API_KEY` from
  `/etc/address-quality/.env.prod` before the local `.env`, so a local run can silently present
  production credentials. The script now refuses any non-loopback target unless
  `AQ_ALLOW_TARGET` is set to that exact URL. Verified both ways: a non-loopback target without
  opt-in aborts at init with no request sent; with opt-in it runs and returns 200.
- **Hard VU cap** of 50, so a mistyped `K6_VUS` cannot turn a diagnostic load test into a stress
  test.

Note on `BASE_URL`: the production fallback does not apply there. With `BASE_URL` unset it
resolves to `http://localhost:7300` (the default), not the production URL. The hazard was
credentials and target confusion, and the guard closes the dangerous direction.

## Limitations

- One run, one host, 4 CPUs, no repeated trials. No confidence interval, no run-to-run variance.
- Five synthetic addresses, round-robin. Not a production traffic mix and not the accuracy dataset.
- Percentiles aggregate ramp-up, hold and ramp-down. No pure steady-state slice was cut.
- Server-side timing is reported as counts and sums only. Percentiles are k6's to own; the
  Prometheus histogram buckets were not used to estimate them.
- k6 and the Go server share 4 CPUs on one host, so client and server contend for CPU. Treat
  absolute numbers as a local diagnostic baseline, not a capacity claim.
- Ran at 11:47 WIB. The project convention is that heavy benchmark runs start at 22:00 WIB. This
  run is light (10 VU, 50 s, loopback) so it ran immediately; anything heavier should be
  scheduled.

## Next

Step 5, benchmark artifact integration: teach the Performance Report to consume
`aq-chore-54-performance-benchmark.json` rather than console text or hard-coded values. Then
aq-chore-52, the Performance Report Audit Page, then end-to-end verification.
