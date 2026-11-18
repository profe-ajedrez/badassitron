# M3-E — Architectural resolution of the imported uint128 backend

Date: 2026-11-17. Baseline: cumulative M3-D source ZIP. **Unreleased source checkpoint.**

## Decision

Keep **two deliberately different boundaries**:

1. `dec128` (M2-E) stores coefficients as imported `github.com/jokruger/dec128/uint128.Uint128` and uses Jokruger primitives for its general coefficient arithmetic. Small, explicit `math/bits` fast paths remain where the decimal engine's historical contract requires them. The public `dec128.New`, `Coefficient`, and `DecodeFromUint128` still accept/return the nominal Badassitron `uint128.Uint128` type. This is a **representation/API boundary**, not an extra arithmetic adapter on every calculation.
2. `dec128/uint128` continues using the **historical, compatible public implementation** by default. `-tags badassitron_m3` retains the complete opt-in Jokruger-backed experimental implementation, but is **not approved as the default**. `dec256`, `dec512`, and `optimus` keep their established public unsigned dependencies. No public alias to the Jokruger type, `unsafe`, or copied upstream implementation is introduced.

This means **Jokruger is already the imported backend for the default 128-bit decimal engine**, without making the historically distinct public unsigned package pay for state/error conversions at every elementary operation. A future upstream/API change may justify reconsidering the unsigned facade, but it is not a prerequisite for M2. Do not describe this as completion of a universal uint128 migration.

## Evidence and rejected alternatives

### Official M3-D comparison provided by the user

The supplied `Archivador(3).zip` contains a **Go 1.27.1 / amd64** eleven-path three-way comparison, and compiler inlining diagnostics. The exact summary files were retained in [m3e-official-evidence](m3e-official-evidence/). These measurements belong to **M3-D**, not newly changed M3-E code.

| Operation | Historical facade | Jokruger direct in M3 build | M3-D facade | M3-D vs historical |
| --- | ---: | ---: | ---: | ---: |
| Add | 0.679 | 0.909 | 1.125 | +65.6% |
| MulExact (Mul64) | 1.339 | 1.323 | 1.565 | +16.9% |
| MulNarrowExact | 2.461 | 2.649 | 3.128 | +27.1% |
| MulHighExact | 2.458 | 2.642 | 3.127 | +27.2% |
| MulOverflow | 1.992 | 1.994 | 2.215 | +11.2% |
| QuoRem | 6.543 | 6.923 | 7.686 | +17.5% |
| WideQuoRem | 9.761 | 9.286 | 10.290 | +5.4% |
| AddOverflow | 0.786 | 0.894 | 1.115 | +41.8% |
| QuoRem64 | 2.621 | 2.627 | 2.647 | +1.0% |
| QuoRemZero | 1.800 | 1.576 | 2.905 | +61.4% |
| WideOverflow | 2.215 | 2.509 | 3.761 | +69.8% |

Nine of eleven paths exceeded the provisional +10% review threshold in that run. This does **not** prove the upstream implementation is universally slower; it distinguishes the direct-backend cost from the facade cost on these exact workloads. Operand construction, compiler and error-path frequency all affect the result.

### Generated call-site evidence

The new [assembly diagnostic](../../tools/migration-m3-architecture/asm.sh) compiles both public-facade modes and examines their symbols. Its **local** Go 1.23.2 output is retained at [m3e-asm-local/calls.md](m3e-asm-local/calls.md). For the opt-in M3 wrapper it observes calls to `github.com/jokruger/dec128/uint128.Uint128.QuoRem` and `QuoRem256By128`. The corresponding historical `QuoRem` is implemented in the facade; its explicit call sites differ. The upstream multiplication wrapper may be inlined while the upstream multiplication method itself is not. Missing symbols are recorded as **not observed**, never as zero cost. This evidence corroborates the existence of an additional delegation boundary; it does not by itself measure the cost of individual instructions or prove a CPU-specific cause.

### Options reviewed

| Option | Decision | Reason |
| --- | --- | --- |
| Alias `badassitron/dec128/uint128.Uint128` to upstream | **Reject** | The distinct exported method sets, error return types and nominal API cannot be preserved with an alias. |
| Make the full M3 adapter the default | **Reject for now** | Validated semantics do not outweigh the official regressions, particularly on frequent operations. |
| Duplicate/copy upstream algorithms into the adapter | **Reject** | Violates the migration's import-only upstream rule and does not remove state-translation obligations. |
| Mix historical and imported arithmetic method-by-method within M3 | **Reject** | Produces an opaque mixed engine and complicates semantics, upgrades and regression attribution. |
| Keep historical public unsigned facade; call Jokruger directly inside M2 coefficient engine | **Adopt** | Satisfies the decimal migration goal and preserves the unsigned API used by wide types without introducing measured M3 overhead. |

## Implemented in M3-E

- Consolidated public `New` and internal `fromParts` through a `newCoefficient(ju.Uint128, ...)` constructor. Previously `fromParts` converted the coefficient to the public unsigned type and called `New` again. The new internal path remains on upstream `uint128`. Public `New` still accepts the exact historical `lu.Uint128` type and converts it once at entry. The six public M2 source files remain exact mirrors of the separately maintained internal candidate.
- Added [4,000-case constructor-boundary regression test](../../dec128/migration_m3e_boundary_test.go): compares public `New` and internal `fromParts` across wide coefficients, scales, positive/negative metadata and out-of-range scales. Preserves the special `New(0,0,true)` normalization and `NewStoredZero(0,true)` raw-sign distinction. Added [paired boundary microbenchmarks](../../dec128/migration_m3e_boundary_bench_test.go).
- Added [architecture guard](../../tools/migration-m3-architecture/main.go), self-tests and [verification script](../../tools/migration-m3-architecture/verify.sh). The guard enforces the public `Uint128{Lo, Hi uint64}` nominal representation, Jokruger coefficient storage in default M2, an explicit coefficient-construction boundary, selected M2 source imports/uses, build constraints, and lack of unreviewed direct unsigned-import changes in wide/optimus sources. Its deliberately limited Go-AST analysis is **not** a substitute for full type checking or exhaustive dynamic-call analysis. The existing independent exported-API and build-selection checks remain mandatory.
- Added [assembly inspection script](../../tools/migration-m3-architecture/asm.sh). It emits only diagnostics; it never silently interprets absence of a symbol as an improvement.
- Preserved previous M3-D source and benchmarks, `go.mod` with `go 1.27` and pinned v1.4.0, and official `go.sum` hashes. No `replace`, `vendor`, or upstream code copied into the delivery.

## Local verification and performance

Local host: **Go 1.23.2, AMD EPYC 9V74, Linux amd64**, using a **temporary copy of the module outside the delivery**, with a test-only `replace` to the supplied **unverified upstream source ZIP** and locally lowered Go directives. These results do **not** authenticate that snapshot as upstream tag v1.4.0 or certify M3-E on Go 1.27.1.

| Check | Result |
| --- | --- |
| Default `go test ./...` + `go vet` | PASS |
| `-tags badassitron_m3` complete tests + vet | PASS |
| Historical decimal and historical+M3 core packages, plus `optimus` | PASS as separate commands; combined full command interrupted by environment limit |
| `optimus`, legacy and legacy+M3 | PASS independently |
| Full default `-race` | PASS |
| M3 `-race`, core and `optimus` | PASS individually; single combined command interrupted by environment limit |
| Legacy+M3 `-race`, core and `optimus` | PASS individually; single combined command not completed |
| Independent Go module consumer (default, M3, legacy+M3) | PASS with temporary local replacements only |
| `go vet` (four build modes) | PASS |
| AST architectural guard and its self-tests | PASS |
| M2 candidate mirror, build selection, uint128 66-symbol API checker | PASS |
| Frozen decimal oracles (default, and M3 separately) | 3,518 + 1,725 each; unexpected differences **0** |
| Independent exact checks (per build) | 389 + 1,634 checks; **0** arithmetic mismatches; skipped/lossy cases still tracked |
| Source-tree links | PASS, 354 local links in 82 Markdown files |

Small same-host *non-interleaved* Go 1.23.2 four-sample 130ms boundary benchmarks (ns/op; zero allocations):

| Boundary | M3-D source median | M3-E source median |
| --- | ---: | ---: |
| Internal `fromParts` | 1.445 | 1.468 |
| Public `New` | 1.795 | 1.483 |

Raw boundary benchmark logs are kept in [m3e-local-validation](m3e-local-validation/). The apparent public-constructor improvement is local and small; it cannot be generalized to Go 1.27, other operands or other callers. This architecture change is principally about removing the **unnecessary type round trip**, not claiming universal speedups.

## Release and continuation gates

M3 is **resolved architecturally, not promoted**. M2-E is the default Jokruger-backed decimal engine. M0 historical authorship/provenance and M7 publication remain open. Only M3-D has the supplied official Go 1.27.1 three-way evidence; M3-E must be verified independently before publication.

On the official Go 1.27.1 host with the published v1.4.0 module, from repository root:

```sh
GOWORK=off bash tools/migration-m2-cutover/verify.sh
GOWORK=off bash tools/migration-m3-architecture/verify.sh
GOWORK=off bash tools/migration-m3-uint128/verify.sh
BENCHTIME=250ms COUNT=7 bash tools/migration-m3-uint128/bench-three-way.sh reports/migration/m3e-official-three-way
GOWORK=off bash tools/migration-m3-architecture/asm.sh reports/migration/m3e-official-asm
```

If official M2-E functional and performance gates pass and provenance is approved, M2 may be published **without** requiring the unsigned facade to switch to M3. Any later migration of public `uint128` requires its own explicit performance decision and source API parity, not an accidental side effect of M2.

## M0–M7 status after M3-E

| Stage | Status |
| --- | --- |
| M0 — Contract and source provenance | Partial. Jokruger published-module evidence has been supplied for prior checkpoints; historical Badassitron source ownership still requires approval. |
| M1 — Bridge and frozen oracles | Implemented; local checks pass on the new boundary. |
| M2 — Default `dec128` | Jokruger-backed by default since M2-E; M3-E local regression tests pass, official Go 1.27.1 rerun open. |
| M3 — Public `uint128` | Architecture decided: historical facade remains default; imported delegate opt-in only. Not a universal full replacement. |
| M4–M5 — Wide types and `optimus` | Not migrated in this stage; functional regressions tested in the configurations specified above. |
| M6 — Performance | Official M3-D comparison preserved; small local M3-E constructor benchmark available; official M3-E comparison open. |
| M7 — Publication | Blocked on official M3-E verification and historical source provenance/release governance. |

## Decisions

- **D61 — ACCEPT:** independent decimal coefficient backend and public unsigned compatibility boundary; do not promote M3.
- **D62 — ACCEPT:** direct upstream-typed internal constructor; keep historical signatures and signed-zero contract.
- **D63 — ACCEPT:** automated source-level architecture guard, API/build-selection cross-checks and assembly diagnostics.
- **D64 — DEFER:** a future default unsigned migration, pending evidence of comparable elementary operations without copying algorithms or changing public signatures.
