#!/usr/bin/env bash
# Compare facade vs directly imported upstream in one untagged build.
set -euo pipefail
cd "$(dirname "$0")/../.."
export GOWORK=off GOMAXPROCS=1
OUT="${1:-reports/migration/m7h-official-backend}"
COUNT="${COUNT:-7}"
BENCHTIME="${BENCHTIME:-250ms}"
mkdir -p "$OUT"
{ go version; go env GOOS GOARCH GOTOOLCHAIN GOFLAGS; grep '^require github.com/jokruger/dec128' go.mod; } > "$OUT/environment.txt"
go test ./dec128/uint128 -run '^$' \
  -bench '^(BenchmarkMigrationM3ThreeWay|BenchmarkMigrationM3FCodec)$' \
  -benchtime="$BENCHTIME" -count="$COUNT" -benchmem > "$OUT/raw.txt"
python3 tools/release/compare_backend.py "$OUT/raw.txt" --samples "$COUNT" --output "$OUT/comparison.md"
