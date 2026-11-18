# M2-C / M6 performance hardening and M0 provenance checkpoint

**Date:** 2026-11-17. **Source of truth:** cumulative
`badassitron_M2B_experimental_public_cutover_20261117.zip`, SHA-256
`ab1ecad4c12050c7e1bdf6a77d974b318b1db1a84eb8c2af3f787f467094f4d6`.
This is an **opt-in candidate checkpoint, not a default cutover**.

## Changes made in this checkpoint

1. **Square root:** `SqrtRoundWithStatus` now calls the imported upstream
   `SqrtRound` only once and evaluates the exactness flag by squaring the
   rounded 128-bit coefficient using imported `uint128.MulCarry`. A common
   decimal power-of-ten alignment compares the 256-bit product with the
   original radicand. All modes, zero underflow and scales 0–19 are covered
   by new independent `math/big` checks and the existing randomized legacy
   comparisons. This removes the second square-root call previously used
   solely to get the inexact flag.
2. **SQL scanning:** `ScanInto` parses strings, bytes and integer driver
   values without an intermediate byte buffer. SQL errors leave the previous
   destination unchanged. The historical value-receiver `Dec128.Scan`
   remains present for interface compatibility. New tests compare normal,
   invalid, NULL and extremal driver values with the historical API.
3. **Exact multiplication:** after an unrepresentable product, the retry
   happens only if `Canonical()` changed at least one operand's scale.
   No repeated full-width multiply occurs when both operands are unchanged.
   Numeric handling of genuine trailing-zero retries remains intact.
4. **Formatting and parsing:** `StringFixed` uses only the imported upstream
   `uint128.StringToBuf` digit primitive and Badassitron-specific sign/scale
   assembly. The common short parser path no longer recomputes the decimal
   position or calls the generic constructor; historical error and signed-zero
   semantics remain covered by oracle tests.
5. **Reproducible M6 gate:** added `bench-m6.sh` and the stdlib-only
   `compare-benchmarks.py`, with three unit tests. It validates matching case
   names, CPU/Go environment and sample counts. It flags >10% sampled median
   slowdown or increased allocations for **review**, not automatic release.
6. **Provenance evidence:** copied the user-provided official **M2-B**
   Go 1.27.1 verification log to this report directory for reproducibility.
   Added a third-party notice register and the exact license/NOTICE texts from
   the **unverified upstream snapshot only** (not Jokruger Go source code).
   The official v1.4.0 NOTICE is still to be checked against the snapshot.
   CI is configured to collect a paired official Go 1.27 M6 run.

All implementation changes are reflected in both
`internal/dec128candidate/*.go` and the six build-tagged `dec128/m2_*.go`
files. The normal build still selects 12 historical implementation files;
`badassitron_m2` selects six M2 files, without an accidental hybrid.

## What was verified versus what was supplied

**Previously supplied verification (M2-B only):** the
[user-supplied Go 1.27.1 log](m2b-user-supplied-go1271-verify-20261117.txt)
reports `go mod verify`, default and tagged full `go test`, full `-race`,
`go vet`, 5,243 frozen cases with zero unexpected differences, independent
mathematical checks, the external module and documentation links as passing.
This log predates **all M2-C source optimizations**. It is not presented as a
fresh official run.

**New local validation (M2-C):** Linux/amd64, AMD EPYC 9V74, Go 1.23.2.
Both variants are compiled against a **separate, unverified upstream ZIP
snapshot** by using a temporary external `-modfile`. The released `go.mod`
is unchanged (`go 1.27`, `github.com/jokruger/dec128 v1.4.0`, no `replace`).
The temporary replacement and upstream code are **excluded** from the ZIP.

The local checkpoints **passed** complete default/M2 tests, both vet runs,
public default/tagged file selection, independent candidate tests, differential
5,243-case replays, exact-check tools and both full `-race` suites. Their
command outputs are retained in this report directory. Results
are local compatibility evidence, **not a Go 1.27.1/v1.4.0 certification**.

## Executed validation matrix for this M2-C tree

| Check | Local result | Evidence |
| --- | --- | --- |
| Default `go test ./... -count=1` | **PASS** | [log](m6c-default-tests-20261117.txt) |
| Tagged `go test -tags badassitron_m2 ./...` | **PASS** | [log](m6c-tagged-tests-20261117.txt) |
| Default + tagged `go vet ./...` | **PASS** | Empty output, executed locally |
| Default complete `go test -race ./...` | **PASS** | [log](m6c-default-race-20261117.txt) |
| Tagged complete `go test -race ./...` | **PASS** | [log](m6c-tagged-race-20261117.txt) |
| Primary differential cases | **3,518/3,518**, 0 differences | [oracle](m6c-primary-oracle-20261117.txt) |
| Supplemental differential cases | **1,725/1,725**, 0 differences | [oracle](m6c-extended-oracle-20261117.txt) |
| Independent primary arithmetic | **389**, 0 mismatches | [exact](m6c-primary-exact-20261117.txt) |
| Independent extended arithmetic | **1,634**, 0 mismatches | [exact](m6c-extended-exact-20261117.txt) |
| Third-party client, both modes | **PASS locally** | [default](m6c-external-default-20261117.txt), [tagged](m6c-external-tagged-20261117.txt) |
| Three benchmark-tool Python unit tests | **PASS** | [log](m6c-tool-tests-20261117.txt) |
| Documentation links | **271 PASS in 60 Markdown files** | Link checker |
| M2 mirror and build-tag isolation | **PASS** | Six M2 mirrors, 12 legacy files never mixed |

The external-module validation used a **temporary separate copy** with Go
1.23 and local upstream replacement, outside the deliverable; official
published-dependency compatibility must still be retested for M2-C.
The 38 explicitly lossy encodes and other non-exact cases skipped by the
independent verifier are **not** counted as mathematical checks.

## Performance evidence

A same-machine paired five-sample comparison of 20 public API operations
was collected after optimizing sqrt, SQL and the short parser. The final
candidate also includes the later multiplication and fixed-string changes;
see the **final** [paired comparison](m6c-final-comparison-20261117.md)
with [default raw samples](m6c-final-default-20261117.txt) and
[M2 raw samples](m6c-final-tagged-20261117.txt).
Final paired medians (default → tagged, ns/op) on this local host:

| Operation | Default | M2-C | Median change |
| --- | ---: | ---: | ---: |
| `Sqrt/FullWidth` | 173.50 | 123.30 | -28.9% |
| `SQL/ScanInto` | 23.64 | 19.67 | -16.8% |
| `Mul/Large` | 21.99 | 16.31 | -25.8% |
| `String/Fixed` | 41.30 | 40.18 | -2.7% |
| `Parse/Long` | 81.21 | 46.71 | -42.5% |

**Three cases require M6 review:** `Add/OppositeSign` (+17.2%), `Div/Exact`
(+14.2%) and `Rounding/DivRound` (+19.3%). Earlier short runs showed
smaller changes; microbenchmark noise and input dependence remain a concern.
These are **not** accepted exceptions. The threshold is only a provisional
**review** trigger for a >10% median regression, not a statistical release
acceptance rule. The core operations were tested with zero heap allocations
(string formatting intentionally allocates the resulting string).

## M0: provenance and license review (partial)

- The published Jokruger v1.4.0 package listing reports MIT.
- The supplied unverified upstream snapshot's `LICENSE` credits Yurii Biurher
  (2026) under MIT; its `NOTICE` contains BSD-3-Clause attribution to Quang
  (2025) and MIT attribution to Luke Champine (2019). The full unaltered texts
  are included under `third_party_licenses/jokruger-dec128-snapshot/`, with an
  explicit warning that snapshot **does not prove tag identity**.
- The source of historical Badassitron code, its chain of authorship and any
  required additional attribution remain a human audit item. No assertion of
  copied source or exclusive authorship is made.
- `go.sum` was empty in the supplied M2-B ZIP. Although the user provided a
  successful official-module verification log for M2-B, its cryptographic
  module hashes were **not contained in that log**; a trusted published
  checksum cannot be reconstructed from an unverified source ZIP. Do not
  invent it. Rerun on a Go 1.27.1 host and commit the generated `go.sum`.
- No source code, local module replacement or vendor tree of Jokruger was
  added to Badassitron. Its licenses and notices are text, not code.

## Release gates and stage status

| Stage | Status | Reason |
| --- | --- | --- |
| M0 | **PARTIAL** | Provenance/history and official notice comparison require review. |
| M1 | **VALIDATED in prior official M2-B run** | M2-C new source still requires official retest. |
| M2 | **OPT-IN IMPLEMENTED; NOT PROMOTED** | Build-tagged public candidate is complete and tested locally. |
| M3 | **PENDING** | Public `uint128` replacement not attempted. |
| M4 | **REGRESSION TESTED** | `dec256`/`dec512` unchanged. |
| M5 | **REGRESSION TESTED** | `optimus` unchanged. |
| M6 | **PARTIAL** | Local 20-case paired matrix; official 1.27.1 workload and remaining exceptions pending. |
| M7 | **BLOCKED** | Do not release before source, licensing, checksums and M6 gates are approved. |

## Required official rerun before cutover

```sh
bash tools/migration-m2-cutover/verify.sh
BENCHTIME=200ms COUNT=5 bash tools/migration-m2-cutover/bench-m6.sh reports/migration/m6-official
```

Inspect the full official `go.sum`, check Jokruger v1.4.0's `LICENSE`/`NOTICE`,
compare external client signatures, inspect benchmark exceptions and document
an explicit M0/M6 release decision. No switch of the default engine was made.
