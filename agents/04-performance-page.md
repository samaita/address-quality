# 04 — Performance page

## Purpose

`tests/api/page/performance.html` is the performance twin of the accuracy
benchmark page (`tests/api/page/benchmark.html`). Same folder, its own metadata
file and its own HTML output. It exists so a performance run can be read and
compared without digging through JSON.

## Files

| File | Role |
| --- | --- |
| `tests/api/page/performance-build.js` | builder: artifact → metadata entry → HTML |
| `tests/api/page/performance-template.html` | page source; `__AQ_DATA__` is replaced with the payload |
| `tests/api/page/performance-metadata.json` | the run history, and the source of truth for before/after |
| `tests/api/page/performance.html` | generated output |
| `Makefile` → `make performance-page` | prompts for a `performance_build` label and the baseline label |

Input is a benchmark artifact (`agents/aq-chore-54-performance-benchmark.json` by
default, `--source` to override). The page never contains hard-coded numbers: it
renders whatever the artifact carries.

## Before and after are explicit, not adjacent

This is the one design point that differs from the accuracy page.

`tests/api/page/build.js` derives "before" from `entries[idx - 1]`, that is, from
file order. `AGENTS.md` warns against exactly that: *"do not compare unrelated
historical entries merely because they are adjacent"*. So the performance builder
never infers a baseline. Each entry records both sides:

```json
{
  "performance_build": "agent-perf-20260927-b-repeat",
  "role": "comparison",
  "before": { "performance_build": "...", "artifact_source": "...", "git_commit": "...", "metrics": { ... } },
  "after":  { "performance_build": "...", "artifact_source": "...", "git_commit": "...", "metrics": { ... } }
}
```

Rules the builder enforces:

* `--before <label>` names the baseline. If other entries exist and `--before`
  is omitted, the run is refused and the known labels are printed.
* The first entry is `role: "baseline"` with `before: null`. A null baseline is
  rendered as `n/a`, never as `0`.
* Re-running an existing `performance_build` label replaces that entry in place
  and keeps its recorded baseline, so a builder fix does not demand re-naming.
* A snapshot carries its own `workload`, `scenario` and `target_environment`.
  When those differ between sides the page raises a "not like-for-like" banner.
  Only scenario *inputs* are compared (`kind`, `max_vus`, `configured_duration_s`);
  `client_iterations` is a result and comparing it would raise the banner on every
  run until it is ignored.

## Reading the page

* **Before and after** — headline cards plus a full metric table with delta and a
  direction verdict (`lower is better` / `higher is better` / `no direction`).
* **Client observed latency** — k6 percentiles, authoritative for what a caller sees.
* **Cold start** — first request on a fresh process, split by stage, kept out of
  the warm percentiles.
* **Server-side stages** — OTel span sums as a share of stage time. Diagnostic
  only, not percentiles.
* **Runtime characteristics**, **not measured**, **limitations** — carried through
  from the artifact. Anything unmeasured stays `not_measured` with its reason.

When both sides share a `git_commit` the page says so and states that the deltas
are run-to-run variance rather than code change. When they differ it warns that a
delta smaller than the spread is not evidence of a change.

## The two recorded runs

Two runs of the identical workload were recorded to seed the history and to bound
the noise floor. Both are loopback, `127.0.0.1:7394`, 10 VUs staged, 50 s.

| | run A (baseline) | run B (repeat) |
| --- | --- | --- |
| label | `agent-perf-20260927-a-baseline` | `agent-perf-20260927-b-repeat` |
| commit | `fa1dd68` | `dc4f3b5` |
| requests | 2814 | 2571 |
| throughput | 56.26 req/s | 51.39 req/s |
| p(50) | 157.34 ms | 165.84 ms |
| p(95) | 269.82 ms | 303.41 ms |
| p(99) | 363.24 ms | 445.86 ms |
| cold start | 5428.0 ms | 5795.7 ms |
| HTTP errors | 0 | 0 |

Between `fa1dd68` and `dc4f3b5` no Go code changed (only benchmark tooling and
docs), so this pair is a direct measurement of run-to-run spread on this host:
throughput moved 8.6%, p(95) moved 12.5%, p(99) moved 22.7%, cold start moved 6.8%.

**Consequence for any future optimization claim: an improvement smaller than
these percentages is not distinguishable from noise on this setup.** Any claimed
win should beat the spread and be repeated.

## How to add a run

1. Run the workload and produce an artifact (see `agents/03-aq-chore-54-k6-benchmark.md`
   for the measurement procedure).
2. `make performance-page`, or directly:

   ```bash
   node tests/api/page/performance-build.js <performance_build> \
     --before <baseline_label> --source <artifact.json>
   ```

3. `performance.html` is regenerated from the last entry in the metadata.

## Derived invariants

* No number is hard-coded in the builder. `warm_requests` and the provenance
  block are read from the run's own evidence (`provenance.txt`, k6 summary) so a
  second run can never inherit the first run's identity or counts.
* k6's `http_req_failed` is a `Rate` where `true` means "request failed": its
  `values.passes` is the error count and `values.fails` is the *non*-error count.
  Reading `fails` as errors reports 2814 errors on a run with zero.
* Trend percentile keys are `p(90)`, `p(95)`, `p(99)` (parenthesised), not `p95`.
* `http.request` is the request total; `stage.*` are its parts. Coverage divides
  the parts by the total and must not include `http.request` as a part.
