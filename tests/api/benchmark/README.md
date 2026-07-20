# Benchmark Test

Measures address-quality engine accuracy by running the test case CSV through
the validation API and recording structured results.

## Usage

```bash
node tests/api/benchmark-test.js
```

### Options (environment variables)

| Variable      | Default                          | Description                     |
|---------------|----------------------------------|---------------------------------|
| `INPUT_FILE`  | `tests/api/cases/address.csv`    | Path to input CSV               |
| `BASE_URL`    | `http://localhost:7300`          | API base URL                    |
| `API_VERSION` | `v1`                             | API version (used in filename)  |

### Examples

```bash
# Default run (address.csv, localhost:7300, v1)
node tests/api/benchmark-test.js

# Custom input file
INPUT_FILE=tests/api/cases/example.csv node tests/api/benchmark-test.js

# Different target
BASE_URL=http://staging:7300 node tests/api/benchmark-test.js
```

## Input CSV Format

| Column | Header       | Maps to                  |
|--------|--------------|--------------------------|
| A      | `Source`     | `source_code` in request |
| B      | `SERP`       | (unused)                 |
| C      | `Address Raw`| `address` in request     |

Row 1 is the header. Data starts from row 2.

## Output CSV Columns

1. Source
2. Address
3. FormattedOutput
4. Province
5. City
6. District
7. SubDistrict
8. PostalCode
9. Confidence
10. NormalizedInput
11. LocationVersion
12. LocationSource
13. AddressID

## Output File

Written to `tests/api/benchmark/YYYY-MM-DD_benchmark_v{version}_{serial}.csv`.

Serial auto-increments per day + version.
