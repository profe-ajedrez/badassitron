#!/usr/bin/env bash
# Source release-candidate technical verification. Requires published Jokruger v1.4.0
# and a Go 1.27.1+ toolchain on the executing host; no source rewrite.
set -euo pipefail
cd "$(dirname "$0")/../.."
export GOWORK=off
export PYTHONDONTWRITEBYTECODE=1

printf 'Go: '; go version
actual="$(go env GOVERSION)"
if [[ ! "$actual" =~ ^go1\.([0-9]+)(\.[0-9]+)? ]]; then
  echo "Cannot parse go version $actual" >&2; exit 1
fi
if (( ${BASH_REMATCH[1]} < 27 )); then
  echo "M7-H verification requires Go 1.27 or newer, got $actual" >&2; exit 1
fi
python3 tools/release/check_backend.py
python3 tools/release/check_hygiene.py
python3 -m unittest discover -s tools/release -p 'test_*.py'
python3 tools/release/check_api.py

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
go mod download -json github.com/jokruger/dec128@v1.4.0 > "$scratch/module-download.json"
go mod verify
python3 tools/release/audit_official.py \
  --download-json "$scratch/module-download.json" \
  --output reports/migration/m7h-official-module-audit.md
# Physical removal of the legacy implementation must not break main tests.
go test ./... -count=1 -timeout=20m
go vet ./...
go test -race -p=2 ./... -count=1 -timeout=20m
# Compatibility alias tags continue to compile; rollback tags fail deliberately.
go test -tags badassitron_m2,badassitron_m3 ./dec128/... ./dec256/... ./dec512/... ./optimus -count=1

for prefix in '' 'extended-'; do
  cases="testdata/migration/oracle/${prefix}cases.jsonl"
  if [[ -z "$prefix" ]]; then
    baseline='testdata/migration/oracle/legacy-results.jsonl'
  else
    baseline="testdata/migration/oracle/${prefix}results.jsonl"
  fi
  go run ./tools/migration-oracle < "$cases" > "$scratch/${prefix}results.jsonl"
  python3 tools/migration-oracle/compare.py "$baseline" "$scratch/${prefix}results.jsonl"
  if [[ -z "$prefix" ]]; then
    python3 tools/migration-oracle/check_exact.py "$cases" "$scratch/${prefix}results.jsonl"
  else
    python3 tools/migration-oracle/check_extended.py "$cases" "$scratch/${prefix}results.jsonl"
  fi
done
# Compile from a separate module; its temporary replace points to this release
# source, never to Jokruger's source. Root go.mod remains unmodified.
external="$scratch/external"
mkdir -p "$external"
cp testdata/migration/external-client/client_test.go "$external/"
cat > "$external/go.mod" <<MOD
module example.net/badassitron-m7-external-fixture

go 1.27

require github.com/profe-ajedrez/badassitron v0.0.0

replace github.com/profe-ajedrez/badassitron => $PWD
MOD
(cd "$external" && go test -mod=mod ./... -count=1)
python3 tools/check_doc_links.py
python3 tools/check_doc_navigation.py
python3 tools/check_example_outputs.py
printf 'Technical verification passed. This is NOT a license/provenance or publishing approval.\n'
