#!/usr/bin/env bash
# Current M7-A architecture guard; only Jokruger-backed production remains.
set -euo pipefail
cd "$(dirname "$0")/../.."
export GOWORK=off
go run ./tools/migration-m3-architecture .
go test ./tools/migration-m3-architecture -count=1
python3 tools/release/check_backend.py
python3 tools/release/check_api.py
