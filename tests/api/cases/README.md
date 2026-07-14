# Test Cases

CSV files used as test data for load/smoke tests against the address validation service.

## Format

| Column      | Description                                                |
| ----------- | ---------------------------------------------------------- |
| `Source`    | Origin of the address (e.g., Scribd, web scrape)           |
| `SERP`      | URL of the source page                                     |
| `Address Raw` | Raw address string to validate (may contain noise, multi-line, etc.) |

## Files

- `example.csv` — A few sample rows showing the expected format.
- `address.csv` — The full private test dataset (git-ignored). Place your actual test data here.
