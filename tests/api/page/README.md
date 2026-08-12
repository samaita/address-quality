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

`tests/api/page/metadata.json` is managed manually. Set at least:

```json
{
  "release": "v0.3.0",
  "build": "42"
}
```

Optional fields: `git_commit`, and `runner` / `server` environment specs (CPU,
memory, OS, k6 version, API version, git commit). Any missing field is shown as
`N/A` on the page. Nothing is fabricated — metrics always come from the JSON
artifacts.

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
3. Update `metadata.json` (release / build / environment)
4. `make benchmark-page`
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
