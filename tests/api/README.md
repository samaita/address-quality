# API Tests (k6)

Load and smoke tests written with [k6](https://grafana.com/docs/k6/latest/).

## Prerequisites

- [k6](https://grafana.com/docs/k6/latest/get-started/installation/) installed and on `PATH`.

## Configuration

Both the k6 scripts (`load-test.js`, `smoke-test.js`) and `benchmark-test.js`
resolve the API base URL and API key through a shared helper (`tests/api/config.js`)
with a consistent precedence:

**BASE_URL** (in order):

1. `K6_BASE_URL` (or `BASE_URL`) environment variable / `-e` flag
2. `K6_BASE_URL` / `BASE_URL` key in `/etc/address-quality/.env.prod`
3. `K6_BASE_URL` / `BASE_URL` key in the project root `.env`
4. Default `http://localhost:7300`

**API_KEY** (in order):

1. `API_KEY` environment variable / `-e` flag
2. `API_KEY` key in `/etc/address-quality/.env.prod`
3. `API_KEY` key in the project root `.env`
4. Blank (no `X-API-Key` header sent)

The scripts log the resolved BASE_URL and whether an API key was configured at
startup, e.g.:

```
BASE_URL: http://localhost:7300 (from default)
API_KEY: set (from .env)
```

The `X-API-Key` header is sent on `/v1/validate` requests when a key resolves;
it is omitted when blank.

## Load Test

The load test (`load-test.js`) ramps up to `K6_VUS` virtual users (default `10`),
spreads a fixed set of Indonesian addresses across them, and POSTs to
`<base-url>/v1/validate`. Each run writes a single JSON summary report to
`tests/api/result/` (`YYYY-MM-DD_<test-name>_<serial>.json`, auto-incrementing
per day) containing the final results: threshold outcomes, per-check breakdown,
and metric aggregates (avg/min/med/max/p90/p95, counts, and rates).

The base URL and `X-API-Key` header follow the [configuration precedence](#configuration)
above. `run-k6.sh` no longer injects `API_KEY`; the k6 scripts resolve it
themselves (k6 exposes system env vars to `__ENV`).

### Local (default base URL)

```bash
make test-api-load
# or directly
./tests/api/run-k6.sh load-test tests/api/load-test.js
```

Targets `http://localhost:7300`.

### Production

```bash
make test-api-load-prod
# or directly
K6_BASE_URL=https://api.samaita.com/address-quality ./tests/api/run-k6.sh load-test tests/api/load-test.js
```

### Tune load

```bash
# 20 VUs against production
K6_VUS=20 make test-api-load-prod
```

## Smoke Test

The smoke test (`smoke-test.js`) runs a single iteration covering `/health`,
valid/invalid address validation, and HTML sanitization.

```bash
make test-api-smoke
# or directly
./tests/api/run-k6.sh smoke-test tests/api/smoke-test.js
```

Target `smoke-test-prod` is not defined; use the env var directly:

```bash
K6_BASE_URL=https://api.samaita.com/address-quality ./tests/api/run-k6.sh smoke-test tests/api/smoke-test.js
```

## Thresholds

Both scripts enforce:

- `http_req_failed`: error rate < 1%
- `checks`: 100% of checks pass

A non-zero exit code from k6 means a threshold was breached; the summary report is still saved to `tests/api/result/`.
