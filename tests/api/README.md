# API Tests (k6)

Load and smoke tests written with [k6](https://grafana.com/docs/k6/latest/).

## Prerequisites

- [k6](https://grafana.com/docs/k6/latest/get-started/installation/) installed and on `PATH`.

## Load Test

The load test (`load-test.js`) ramps up to `K6_VUS` virtual users (default `10`),
spreads a fixed set of Indonesian addresses across them, and POSTs to
`<base-url>/v1/validate`. Result CSVs are written to `tests/api/result/`
(`YYYY-MM-DD_<test-name>_<serial>.csv`, auto-incrementing per day).

The `X-API-Key` header is sent using the `API_KEY` value from the root `.env`
(`run-k6.sh` loads it; an `API_KEY` already set in the environment wins).
The header is omitted when the key is empty.

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

A non-zero exit code from k6 means a threshold was breached; results are still saved to `tests/api/result/`.
