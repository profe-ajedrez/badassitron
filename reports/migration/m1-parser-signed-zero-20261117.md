# M1 continuation — parser optimization and stored-zero fidelity

**Date:** 2026-11-17  
**Scope:** experimental `internal/dec128bridge`; **public `dec128` not switched**.  
**Starting authority:** `badassitron_M1_independent_codecs(1).zip` (SHA-256 `dcb36fe9d25a1ea48309c67c4192fdb78907b86d0dad62935e7b9db4c89f1a64`).  
**Reference-only upstream snapshot:** `dec128-main(1).zip` (SHA-256 `c93593087448a3d888e887fbfcbc6a98abfb9d9ae8484a1e7bb44c38d9101d3b`).

## Implemented

1. Reworked `ParseCompatible` to accumulate small decimal coefficients in one validated pass. For long decimal strings, it uses the imported Jokruger `uint128.FromString`, `Mul64`, and `Add64` directly on original integer/fraction substrings; the previous temporary concatenation and a redundant full-character validation pass are gone. The short/long split still uses the **original input length**, preserving sign-sensitive boundary quirks. No upstream sources were copied.
2. Preserved parser error precedence and unusual historical forms: empty input as zero, short trailing point (`5.`) accepted, long trailing point rejected, out-of-range fractional scale checked before coefficient parsing in the long path, overflow preceding malformed trailing characters, and negative all-zero *long integer* retaining a stored sign at scale zero. Invalid exponents remain rejected.
3. Added independent `Compatible` read-only accessors (`Coefficient`, `Precision`, `StoredNegative`, `IsZero`, `ErrorDetails`) for differential checks without converting the result through the historical public type. Restored the historical stored-zero sign on `WrapLegacy(...).Legacy()` even when the signed zero has scale zero (reachable from a legacy rescale, although `legacy.New(0,0,true)` normalizes it). This migration-only interop helper still uses a legacy operation for that special constructor case.
4. Fixed `RescaleRoundStrict` to distinguish (a) an already-zero operand retaining its stored sign at the output scale and (b) a nonzero negative operand rounded to zero, for which the historical quantizer normalizes the output sign. Kept all seven historical rounding modes and the `inexact` bit.
5. Fixed `AddStrict` / `SubStrict` zero results to preserve stored scale and sign when the historical same-sign calculation does, while retaining canonical zero for cancellation. The equal-scale fast path no longer conflates `IsNegative()` with `StoredNegative()` for signed-zero operands.
6. Added `parser_gates_test.go` (long-input boundaries, deterministic 30,000-case differential, seeded 10,000-case `math/big` coefficient oracle, scale-zero signed zero) and `signed_zero_migration_test.go` (20×20×7 rescale grid, quantized-to-zero tests, arithmetic/rounding signed-zero cross product). Extended codec benchmarks over six input classes.

## Verified locally — with important caveats

Local environment: **linux/amd64, AMD EPYC 9V74 80-Core Processor, Go 1.23.2**. Tests temporarily used a `go 1.23` directive in both working modules and a `replace github.com/jokruger/dec128 => ../dec128-main`. These changes are **not in the delivered `go.mod`**. Consequently these results do **not** certify Go 1.27 or the exact upstream `v1.4.0` release.

| Check | Executed result | Supporting log |
| --- | --- | --- |
| `go test ./... -count=1` | PASS, all packages (including `dec128`, `dec256`, `dec512`, `optimus`) | [suite](m1-parser-tests-20261117.txt) |
| `go vet ./...` | PASS | [vet](m1-parser-vet-20261117.txt) |
| `go test -race` without `optimus` | PASS | [race subset](m1-parser-race-except-optimus-20261117.txt) |
| Selected `optimus` tests with `-race` | PASS | [race selected](m1-parser-race-optimus-selected-20261117.txt) |
| `go test -race ./...` | **INCOMPLETE**, environment execution timed out; no full PASS claimed | [partial log](m1-parser-race-full-interrupted-20261117.txt) |
| 3,518-case frozen primary oracle | PASS, 0 unexpected/approved differences | [compare](m1-parser-oracle-primary-20261117.txt) |
| Primary independent Decimal arithmetic checks | PASS, 389 checked / 159 skipped, 0 mismatches | [exact](m1-parser-oracle-exact-20261117.txt) |
| 1,725-case frozen extended oracle | PASS, 0 differences | [compare](m1-parser-oracle-extended-20261117.txt) |
| Independent extended checks | PASS, 1,634 checks / 563 skipped / 38 lossy encodes / 0 mismatches | [exact](m1-parser-oracle-extended-exact-20261117.txt) |
| Example output checks | PASS, 11 examples | [examples](m1-parser-examples-20261117.txt) |
| Python comparator unit tests | PASS, 6 tests | Invoked `python3 -m unittest discover -s tools/migration-oracle -p test_*.py` |
| Exported API snapshot | Existing additive `Dec128.StoredNegative()` only; 0 removed/changed declarations | AST inventory comparison with `testdata/migration/api-baseline.json` |

**Oracle scope limitation:** the frozen 5,243-case streams run against the still-legacy **public** `dec128`. They confirm unchanged public behavior and fixture reproducibility, not completion of the future public M2 replacement. The new experimental parser and arithmetic routes have their own dedicated differential, signed-zero, and independent `math/big` tests. No frozen oracle results were rewritten.

## Parser performance

Local paired microbenchmarks were repeated five times with `GOMAXPROCS=1`, `-benchmem`, Go 1.23.2, on the same host. Exact commands and raw samples: [before](m1-parser-before-20261117.txt), [after](m1-parser-after-20261117.txt). Medians are in ns/op; both implementations allocated **0 B/op, 0 allocs/op** in every listed case.

| Input class | Historical | New bridge | Observation |
| --- | ---: | ---: | --- |
| Main long fractional input, **before bridge change** | 50.87 | 103.20 | Starting regression against historical |
| Main long fractional input, **after bridge change** | 48.15 | 43.71 | Bridge ~58% faster than its own starting value; ~9% faster than historical in this run |
| Short integer | 20.70 | 22.21 | Bridge ~7% slower |
| Short fractional | 12.76 | 14.18 | Bridge ~11% slower |
| Long fractional (shape suite) | 41.91 | 34.20 | Bridge ~18% faster |
| Long integer (near max coefficient) | 109.00 | 57.15 | Bridge ~48% faster |
| Long zero integer | 124.00 | 70.67 | Bridge ~43% faster |
| Overlong fraction scale | 4.57 | 5.97 | Bridge ~31% slower on early rejection |

The before/after parser comparison spans separate runs in this environment; do not treat these ratios as a controlled multi-platform performance guarantee or a replacement for full M6 work. The shape benchmarks are **after-only** comparisons to the historical implementation, not measurements of the prior bridge on those shapes. Long-input improvements must not be generalized to short inputs.

## Compatibility decisions and unresolved issues

- No public signatures, import paths, error enum values, or `go.mod` dependency requirement changed. The preexisting **additive** `StoredNegative()` method is the single difference from the M0 AST snapshot.
- Kept `WrapLegacy`/`Legacy` as compatibility-only gateways and `legacy` input/output in experimental arithmetic operations. `RescaleRoundStrict` now uses a migration-only legacy zero constructor for an otherwise unconstructible signed-zero state; this is **not** a claim that production arithmetic is independent of the legacy type.
- Parsing and normal codec/SQL paths no longer call legacy parse/format/SQL functions. Invalid input and errors are still mapped through imported Jokruger state and the existing legacy error vocabulary.
- The GitHub Go package registry publicly lists `github.com/jokruger/dec128@v1.4.0`, but local Go module download, module checksum generation and Go 1.27 acquisition were blocked by DNS/network restrictions. The uploaded `main` snapshot is **not** verified to match that tag. The delivered `go.sum` is empty; **official reproducibility remains blocked**.
- The `go test -race ./...` process exceeded the local execution limit while `optimus` was outstanding; the subset passes are not substituted for the full gate.
- M1 cannot close until the bridge's production paths can run without legacy decimal arithmetic, the upstream published tag and Go 1.27 gates can run, and a standalone candidate (M2) can pass all frozen behavior checks. M2 is **not started** in this checkpoint: partial public cutover would violate the atomic-migration rule.

## Stage status (M0–M7)

| Stage | Status after this checkpoint | Next gate |
| --- | --- | --- |
| M0 — Contract | **Partial** | Final source-provenance/license audit and official toolchain validation |
| M1 — Differential oracle, arithmetic bridge, compatible codecs | **Partial / advanced** | Official dependency and toolchain verification; eliminate remaining normal legacy interop and close full race/performance exceptions |
| M2 — Atomic public `dec128` | **Not started / gated** | A complete separate facade candidate with all API methods, differential oracle, and no hybrid public release |
| M3 — `uint128` primitives | **Pending** | Start only after validated M2 |
| M4 — Wide decimals | **Pending** | Dedicated 256/512 regressions and benchmarks after M3 |
| M5 — `optimus` integration | **Pending** | Cross-tier exactness/promotion tests against migrated facade |
| M6 — Full performance and documentation | **Partial evidence only** | Whole-library, same-toolchain historical/upstream/candidate runs |
| M7 — Release audit | **Blocked** | All upstream, API, licensing and Go 1.27 release gates |

**No upstream code, local dependency directory, or `replace` directive may be distributed in the final ZIP.**
