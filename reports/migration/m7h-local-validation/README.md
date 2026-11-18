# M7-H local diagnostic validation (NOT official)

**Date:** 2026-11-18. Local host: Go **1.23.2**, Linux amd64, shared CPU.
This is a validation of an isolated **disposable copy** of this source tree,
not a Go 1.27.1 verification and not a test using authenticated upstream
Jokruger v1.4.0.

## Diagnostic dependency setup (outside the delivered repository)

The supplied `dec128-main(2).zip` declares `module github.com/jokruger/dec128`
and originally requires `go 1.24`. Its archive SHA-256 is
`c93593087448a3d888e887fbfcbc6a98abfb9d9ae8484a1e7bb44c38d9101d3b`.
It has not been authenticated as the published tag **v1.4.0**.

Only in separate temporary module copies, the `go` directives were lowered to
`go 1.23`, and the temporary Badassitron `go.mod` included a local upstream
`replace`. The release `go.mod` and `go.sum` were not changed. Tests ran with
`GOWORK=off`, `GOTOOLCHAIN=local`, `GOPROXY=off` and `GOSUMDB=off`.

## Commands executed on the temporary copy

```sh
GOWORK=off go test ./... -count=1 -timeout=20m
GOWORK=off go vet ./...
GOWORK=off go test -race -p=2 ./... -count=1 -timeout=20m
GOWORK=off go test -tags badassitron_m2,badassitron_m3 ./dec128/... ./dec256/... ./dec512/... ./optimus -count=1
GOWORK=off go run ./tools/migration-oracle < testdata/migration/oracle/cases.jsonl > primary.jsonl
python3 tools/migration-oracle/compare.py testdata/migration/oracle/legacy-results.jsonl primary.jsonl
python3 tools/migration-oracle/check_exact.py testdata/migration/oracle/cases.jsonl primary.jsonl
GOWORK=off go run ./tools/migration-oracle < testdata/migration/oracle/extended-cases.jsonl > extended.jsonl
python3 tools/migration-oracle/compare.py testdata/migration/oracle/extended-results.jsonl extended.jsonl
python3 tools/migration-oracle/check_extended.py testdata/migration/oracle/extended-cases.jsonl extended.jsonl
# Each of examples/01_... through examples/11_... was run using go run ./examples/<directory>.
```

Results: all commands above passed. The two frozen oracle datasets contain
**3,518 + 1,725** cases, both with **0 unexpected differences**. Independent
exact checks passed. All eleven examples exited zero. Each corresponding
`*.log` and `*.exit` file here preserves output and the actual exit status.
Generated full oracle outputs are reproducible from bundled frozen inputs;
checksums are in `generated-oracle-sha256.txt`.

## Official technical gate deliberately not counted as passed

`official_gate_go1232_rejected.*`: `GOTOOLCHAIN=local` made the release
gate reject installed Go 1.23.2 (exit **1**). With default automatic toolchain
selection, `official_gate_unavailable.*` captured a failed download of Go
1.27 (also exit **1**). Neither attempt performed the official test matrix.

Required next step: run `GOWORK=off bash tools/release/verify.sh` on the exact
candidate using Go 1.27.1+ and **official** Jokruger v1.4.0; attach logs
and candidate checksum. Even a pass does not grant legal provenance clearance.
