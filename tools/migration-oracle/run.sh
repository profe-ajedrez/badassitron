#!/usr/bin/env bash
# Independently execute the same cases against two module roots and compare.
set -euo pipefail
if [[ $# -lt 2 || $# -gt 3 ]]; then
  echo 'usage: run.sh LEGACY_MODULE_ROOT CANDIDATE_MODULE_ROOT [APPROVED_DIFFS_JSON]' >&2
  exit 2
fi
legacy=$(cd "$1" && pwd)
candidate=$(cd "$2" && pwd)
script_dir=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$script_dir/../.." && pwd)
casefile="$repo/testdata/migration/oracle/cases.jsonl"
out=$(mktemp -d)
trap 'rm -rf "$out"' EXIT
(
 cd "$legacy"
 GOWORK=off go run "$script_dir/main.go" < "$casefile" > "$out/legacy.jsonl"
)
(
 cd "$candidate"
 GOWORK=off go run "$script_dir/main.go" < "$casefile" > "$out/candidate.jsonl"
)
if [[ $# == 3 ]]; then
 python3 "$script_dir/compare.py" "$out/legacy.jsonl" "$out/candidate.jsonl" --approved "$3"
else
 python3 "$script_dir/compare.py" "$out/legacy.jsonl" "$out/candidate.jsonl"
fi
