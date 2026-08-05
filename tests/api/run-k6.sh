#!/usr/bin/env bash
set -euo pipefail

TEST_NAME="$1"
SCRIPT="$2"
RESULT_DIR="tests/api/result"
DATE=$(date +%Y-%m-%d)

mkdir -p "$RESULT_DIR"

API_KEY="${API_KEY:-}"
if [ -z "$API_KEY" ]; then
  ENV_FILE="$(cd "$(dirname "$0")/../.." && pwd)/.env"
  if [ -f "$ENV_FILE" ]; then
    API_KEY="$(sed -n 's/^API_KEY=//p' "$ENV_FILE" | tail -1)"
  fi
fi

PREFIX="${DATE}_${TEST_NAME}_"
LATEST=$(find "$RESULT_DIR" -maxdepth 1 -name "${PREFIX}*.csv" 2>/dev/null | \
  sed -n "s/.*_\([0-9]\{4\}\)\.csv$/\1/p" | sort -n | tail -1)

if [ -z "$LATEST" ]; then
  NEXT="0000"
else
  NEXT=$(printf "%04d" $((10#$LATEST + 1)))
fi

OUTPUT="${RESULT_DIR}/${PREFIX}${NEXT}.csv"
k6 run --out "csv=${OUTPUT}" -e "API_KEY=${API_KEY}" "$SCRIPT"
