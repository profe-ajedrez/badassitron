#!/usr/bin/env bash
# Documented repro only; do NOT run on a machine without official Go 1.27.1.
# All candidate trees are separate copies, with no go.mod edits or replace.
set -euo pipefail
cd "$(dirname "$0")/../../.."
export GOWORK=off GOTOOLCHAIN=local
[[ "$(go version)" =~ go1\.27\.1[[:space:]] ]] || { echo 'requires exactly Go 1.27.1' >&2; exit 1; }
python3 tools/release/m7g-stage-variant.py --baseline . --variant m7e --output /tmp/m7g-e
python3 tools/release/m7g-stage-variant.py --baseline . --variant g2 --output /tmp/m7g-g2
python3 tools/release/m7g-stage-variant.py --baseline . --variant g3 --output /tmp/m7g-g3
# Full acceptance checks: supply candidate folder instead of M7-E ZIP.
# Go 1.27.1 must download the official github.com/jokruger/dec128@v1.4.0.
# For the most plausible candidate (g2), e.g.:
python3 tools/release/bench-interleaved.py --baseline . --candidate /tmp/m7g-g2 --output reports/migration/m7g-official-interleaved-g2 --blocks 8 --benchtime 250ms
GOWORK=off go test ./... -count=1
GOWORK=off go test -race ./... -count=1
GOWORK=off go vet ./...
GOWORK=off bash tools/release/verify.sh
# Inspection (the Go compiler may inline instantiations and omit a named symbol):
GOWORK=off go test -c -o /tmp/m7g-uint128.test ./dec128/uint128
GOWORK=off go test -run '^$' -gcflags='github.com/profe-ajedrez/badassitron/dec128/uint128=-m=2' ./dec128/uint128 2> /tmp/m7g-inline.txt
GOWORK=off go tool objdump -s 'github.com/profe-ajedrez/badassitron/dec128/uint128.*FromString' /tmp/m7g-uint128.test > /tmp/m7g-fromstring-objdump.txt
