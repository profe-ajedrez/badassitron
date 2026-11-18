# M7-C — imported unsigned adapter performance

Date: 2026-11-18. Status: **local candidate; NOT approved for publication**.

## Scope and invariants

This is an incremental optimization of M7-A, not a new arithmetic engine.
`dec128/uint128/m3_adapter.go` continues to delegate all 128-bit arithmetic
through the imported `github.com/jokruger/dec128/uint128 v1.4.0`. The public
nominal `Uint128{Lo, Hi}` type, exported declarations, error codes and zero
results on arithmetic errors are unchanged. No upstream algorithms are
copied, and no `replace` or `vendor` is shipped. The historical arithmetic
engine remains absent.

## Changes

- `Add`, `Add64`, `Mul` and `Mul64`: use the upstream `state.OK=0` /
  `state.Overflow=5` to map directly to Badassitron `errors.None=0` /
  `errors.Overflow=3` through `(uint8(s)&1)*3`. Jokruger already returns
  a zero result for overflow; no additional branch or reconstruction.
- `Sub`, `Sub64`: upstream `state.Underflow=6` maps to the legacy
  `errors.Underflow=4` using `uint8(s)&4`.
- `Div`, `Mod`, `QuoRem`, and 64-bit variants: upstream ordinary quotient
  status `0/4` maps to `0/2` by shifting one bit. Removed the duplicate
  zero-divisor check from `QuoRem`; the imported implementation performs it.
- `QuoRem256By128`, `QuoRem192By64`: upstream states `OK=0`,
  `DivisionByZero=4`, `Overflow=5` map to `None=0`, `DivisionByZero=2`,
  `Overflow=3` through `(uint8(s)+1)>>1`. Removed the duplicate wide
  quotient overflow precheck; the imported implementation performs it.
- Replaced *only* the wrapper-level state mapping. The full, general
  `fromState` mapping remains for parsing and other upstream states.
- Added release-contract tests and an A/B benchmark comparator.

The branchless formulas deliberately depend on published **v1.4.0** state
numbers; they are not a general state conversion function. The tests fail if
these constants change. Upgrading Jokruger requires reviewing the assumptions.

## Local verification completed

Environment: container Go **1.23.2** on Linux amd64. To execute the tests in
this environment only, a scratch copy changed `go.mod` to `go 1.23` and used a
local `replace` pointing to the supplied Jokruger source snapshot. These
changes do **not** appear in the delivery. This does not substitute for
Go 1.27.1 with the official versioned module.

- `TestReleaseM7CStateEncoding` PASS: exact numeric statuses.
- `TestReleaseM7CImportedArithmeticAndErrors` PASS: 12,000 deterministic
  operand pairs, covering narrow and wide division and success/error paths.
- Existing `uint128` correctness and `math/big` oracles PASS.
- `go test ./... -count=1`, `go vet ./...`, and compatible alias-tagged
  decimal/consumer suites PASS.
- `go test -race -p=2 ./...` was interrupted by host execution timeout.
  Every package reported before interruption passed; `optimus` and remaining
  packages passed in separately completed race invocations. This is package
  coverage, **not a complete global `-race` run**.
- 3,518 main and 1,725 extended frozen oracle cases PASS, zero unexpected
  differences; independent `math/big` checks report no mismatches.
- External-client fixture PASS with temporary scratch-only replacements.
- 19 release-tool unit tests, backend guard and frozen public API guard PASS;
  1,030 historical identifiers and all 66 `uint128` signatures retained.
- [Raw local A/B evidence](m7c-local/quick-ab-comparison.md) and
  [scope of verification](m7c-local/validation-summary.txt) are included.

## Performance observations (NOT official)

The local A/B run used the same Go 1.23.2 and separate compilation of M7-A
and M7-C, each also calling direct Jokruger in the same executable. Three
samples of 65ms per route are too short for a performance acceptance gate.
The percentages here are within-build overhead relative to direct upstream:

| Operation | M7-A local | M7-C local |
|---|---:|---:|
| `Add` | +57.5% | -2.0% |
| `MulExact` | +9.5% | +0.7% |
| `MulOverflow` | +4.9% | -3.1% |
| `QuoRem` | +7.6% | -23.2% |
| `WideQuoRem` | +11.9% | -5.8% |

A separate five-sample M7-C trial found `Add` at **+2.1%** and `MulExact`
at **+2.2%**. The large negative `QuoRem` percentage is not interpreted
as a genuine speedup over Jokruger's algorithm: these are different compiled
call paths under different optimization contexts, and the host is noisy.
The compiler report is more structurally useful than a noisy percentage:

| Wrapper | M7-A Go 1.23.2 | M7-C Go 1.23.2 |
|---|---|---|
| `Add` | inline, cost 60 | inline, cost 57 |
| `Mul` | inline, cost 77 | inline, cost 74 |
| `QuoRem` | not inline, cost 96 | inline, cost 80 |
| `QuoRem256By128` | not inline, cost 158 | inline, cost 80 |

[Baseline compiler evidence](m7c-local/inlining-baseline.txt) and
[optimized compiler evidence](m7c-local/inlining-optimized.txt) are included.
Inlining of the wrapper does not imply inlining of Jokruger's arithmetic;
this does **not** establish what Go 1.27.1 will emit.

The prior **official M7-A** Go 1.27.1 benchmark reported `Add` +27.1%,
`MulExact` +16.6%, `QuoRem` +13.4%, and `WideQuoRem` +16.7% over direct
Jokruger. No official M7-C timing is claimed.

## Release decision and pending gates

M7-C is a candidate optimization, not a published or validated release.
It must pass `GOWORK=off bash tools/release/verify.sh` **on this exact source
with Go 1.27.1 and the official Jokruger v1.4.0 module**. Run
`BENCHTIME=250ms COUNT=7 bash tools/release/bench-backend.sh
reports/migration/m7c-official-backend` and compare with official M7-A
through `tools/release/compare-iterations.py`; also inspect `optimus`,
`dec256` and `dec512` consumer regressions before claiming a universal
5% overhead goal.

Historical Badassitron source provenance and redistribution clearance remain
independent publication blockers; a passing technical gate is not a license
or publication approval.
