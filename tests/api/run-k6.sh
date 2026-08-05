#!/usr/bin/env bash
set -euo pipefail

TEST_NAME="$1"
SCRIPT="$2"
RESULT_DIR="tests/api/result"
DATE=$(date +%Y-%m-%d)

mkdir -p "$RESULT_DIR"

PREFIX="${DATE}_${TEST_NAME}_"
LATEST=$(find "$RESULT_DIR" -maxdepth 1 -name "${PREFIX}*.csv" 2>/dev/null | \
  sed -n "s/.*_\([0-9]\{4\}\)\.csv$/\1/p" | sort -n | tail -1)

if [ -z "$LATEST" ]; then
  NEXT="0000"
else
  NEXT=$(printf "%04d" $((10#$LATEST + 1)))
fi

OUTPUT="${RESULT_DIR}/${PREFIX}${NEXT}.csv"
k6 run --out "csv=${OUTPUT}" "$SCRIPT"
