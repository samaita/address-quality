# Benchmark Page

Self-contained, embeddable page that reports one Address Quality release/build:
accuracy, performance and reproducibility metadata.

The only artifact is **`benchmark.html`** — a single HTML file with the data inlined.
It renders standalone and inside an `<iframe>`, with no external dependencies, no
CORS, and scoped CSS that cannot affect the host page.

The build also writes **`full-benchmark.html`** — the same page but untrimmed:
it keeps the raw address input for every record, renders the top-100 failed
matches ranked by confidence (high to low) and always shows the performance run.
Because it embeds confidential raw inputs, `full-benchmark.html` is gitignored
and intended only for local inspection, never for publishing.

## Build

```bash
make benchmark-page
```

`tests/api/page/build.js` reads the **latest** benchmark and load-test JSON from:

- `tests/api/benchmark/YYYY-MM-DD_benchmark_v1_0000.json`
- `tests/api/result/YYYY-MM-DD_load-test_0000.json`

and writes `tests/api/page/benchmark.html`.

## Release / build metadata

Two files live in `tests/api/page`:

- **`release.json`** — the current run's manually-managed release/build/env
  config. Set at least:

  ```json
  {
    "release": "v0.3.0",
    "build": "42"
  }
  ```

  Optional fields: `git_commit`, and `runner` / `server` environment specs (CPU,
  memory, OS, k6 version, API version, git commit). Any missing field is shown
  as `N/A` on the page. Nothing is fabricated — metrics always come from the
  JSON artifacts.

- **`metadata.json`** — an auto-appended history array with one entry per test
  run (keyed by `release` + `build`). Each entry stores the full release table
  (release, build, git commit, benchmark timestamp, dataset version/source,
  dataset generation timestamp, load-test timestamp, benchmark source, benchmark
  runner, runner/server env) plus the computed run metrics:

  ```json
  {
    "release": "v0.3.0",
    "build": "42",
    "git_commit": "02ba28d",
    "benchmark_timestamp": "2026-08-11",
    "dataset_version": "2025",
    "dataset_source": "kemendagri",
    "dataset_generation_timestamp": null,
    "load_test_timestamp": "2026-08-06T15:23:05.382Z",
    "benchmark_source": "2026-08-11_benchmark_v1_0000.json",
    "benchmark_runner": null,
    "runner": {},
    "server": {},
    "metrics": {
      "overall_accuracy": 49.1,
      "dataset_size": 106,
      "average_confidence": 72.8,
      "dataset_version": "2025",
      "hierarchy": [
        { "level": "Province", "correct": 87, "pct": 82.1 },
        { "level": "City", "correct": 81, "pct": 76.4 },
        { "level": "District (Kecamatan)", "correct": 76, "pct": 71.7 },
        { "level": "Subdistrict (Kelurahan)", "correct": 55, "pct": 51.9 },
        { "level": "All four levels (exact)", "correct": 52, "pct": 49.1 }
      ]
    }
  }
  ```

  `build.js` computes the metrics from the latest benchmark JSON and appends the
  entry on every run. If an entry with the same `release` + `build` already
  exists it is replaced in place instead of duplicated.

## Before / After comparison

`metadata.json` keeps every run, so `build.js` passes the previous entry (the
run immediately before the current one) as `before` in the page payload. The
page renders Overall accuracy, Dataset size, Average confidence and the
Correct-by-hierarchy-level table with **Before** / **After** / **Δ** columns so
regressions and improvements across consecutive runs are visible. When only one
run exists, the Before values show `N/A`.

## Privacy

Raw address inputs are confidential. They are embedded in `benchmark.html`
**only** for the three Top-Challenge example records. All other records are
embedded trimmed (Address ID, status, confidence, formatted location, ground
truth and derived flags). Failure tables and lists never include raw inputs.

`full-benchmark.html` is the deliberate exception: it is the untrimmed build used
to inspect raw inputs against the resolved/actual addresses locally. It is
excluded from git (`.gitignore`) and must not be published.

## Updating after a new benchmark run

1. Run the benchmark: `make benchmark` (writes `tests/api/benchmark/…json`)
2. Run the load test: `make test-api-load` (writes `tests/api/result/…json`)
3. Update `release.json` (release / build / environment)
4. `make benchmark-page` (appends the new run to `metadata.json`)
5. Publish the new `benchmark.html`

## Embedding from the external website

Serve `benchmark.html` from any static host (or copy it into the external
website) and embed it:

```html
<iframe
  src="https://samaita.com/benchmark.html"
  title="Address Quality Benchmark"
  width="100%"
  height="1200"
  style="border:0"
></iframe>
```

The page is fully self-contained; no `allow-same-origin` or `allow-scripts`
sandbox flags are required for the HTML itself, but keep them set if your
sandbox policy demands it.
