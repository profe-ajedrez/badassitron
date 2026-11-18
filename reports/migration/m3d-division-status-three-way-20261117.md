# M3-D — Operation-specific division status mapping and 11-path backend comparison

Date: 2026-11-17. Source authority: cumulative M3-C archive. **Experimental, not released.**

## Scope and actual implementation

The public `dec128` engine remains the Jokruger-backed M2-E default. The historical `dec128/uint128` implementation remains the default for unsigned integers. `-tags badassitron_m3` still selects imported Jokruger `uint128` arithmetic while preserving Badassitron's nominal `Uint128`, its public `Lo`/`Hi` limbs, exported signatures, and historical error codes.

Changes in `dec128/uint128/m3_adapter.go` are deliberately limited to the **ordinary unsigned division family** (`Div`, `Div64`, `Mod`, `Mod64`, `QuoRem`, `QuoRem64`): convert operands/results by Go value conversion and translate the two states returned by upstream v1.4.0, `state.OK` and `state.DivisionByZero`, directly. `QuoRem` and `QuoRem64` explicitly retain zero quotient/remainder on error. They still execute imported Jokruger arithmetic. **Wide division (`QuoRem256By128`, `QuoRem192By64`) continues to use the full state translator**, because its contract also includes overflow. No imported source was copied and no `unsafe` was introduced.

The mapping depends on Jokruger's pinned unsigned-division contract and has new tests asserting the expected states. It must be reviewed upon any upstream upgrade; it is **not** a universal state mapping for decimals or wide division.

## Additional validation and benchmark path coverage

- New `migration_m3_division_status_test.go`: **2,000 deterministic input sets** covering `Div`, `Div64`, `Mod`, `Mod64`, `QuoRem`, `QuoRem64`, `QuoRem256By128`, and `QuoRem192By64`. Compares value, remainder, and mapped error to direct imported Jokruger, across valid divisors, zero divisors, and overflow cases. PASS on both historical and M3 builds. This is an interop-state test, **not** an independent arithmetic proof; existing `math/big` tests remain separate.
- Three-way benchmarks now include `AddOverflow`, `QuoRem64`, `QuoRemZero`, and `WideOverflow`, in addition to the prior seven. This prevents a successful-division measurement from obscuring error-path overhead. Both facade and direct backend paths retain results and status in sinks; inputs vary in a 32-element array.
- The original M3-C seven-operation measurements remain preserved as historical evidence. A trial dispatching narrow `Mul` to the imported `Mul64` primitive improved the exact narrow case but made overflowing `Mul` substantially slower, because the wrapper lost inlining. **That experiment was reverted.** No multiplication algorithm changed.

## Local comparison (not official)

Go 1.23.2, AMD EPYC 9V74, `GOMAXPROCS=1`, supplied upstream source ZIP used *only as a temporary module replacement outside the delivered tree*, three samples at 95ms per route, nanoseconds per operation. M3-D is compared to the historical facade and direct Jokruger under the same test harness. All measured allocations were zero.

| Operation | Historical facade | Jokruger direct in M3 build | M3-D facade | M3-D vs historical |
| --- | ---: | ---: | ---: | ---: |
| Add | 1.146 | 1.176 | 1.528 | +33.3% |
| MulExact (Mul64) | 1.680 | 1.776 | 1.985 | +18.2% |
| MulNarrowExact (Mul) | 2.563 | 2.963 | 3.222 | +25.7% |
| MulHighExact (Mul) | 2.861 | 3.090 | 3.207 | +12.1% |
| MulOverflow | 2.289 | 2.321 | 2.687 | +17.4% |
| QuoRem | 10.920 | 10.220 | 10.300 | -5.7% |
| WideQuoRem | 11.450 | 11.680 | 11.960 | +4.5% |
| AddOverflow | 1.389 | 1.225 | 1.356 | -2.4% |
| QuoRem64 | 4.025 | 3.500 | 3.539 | -12.1% |
| QuoRemZero | 1.887 | 2.337 | 3.258 | +72.7% |
| WideOverflow | 2.700 | 2.619 | 3.831 | +41.9% |

This small benchmark sample is **diagnostic**, not a release decision. The error paths remain materially slower at the facade boundary; `Mul` success and `Add` remain above the provisional 10% threshold. In a separately executed sequential 5×120ms local test of the quotient subset, M3-C `QuoRem/Facade` had a median of 11.290 ns/op and M3-D 9.892 ns/op; these are not interleaved samples and do not establish a stable 12.4% improvement. Raw M3-D benchmark and compiler diagnostics are in `reports/migration/m3d-local-three-way/`.

## Local verification status

| Check | Result |
| --- | --- |
| `go test -p=2 ./... -count=1` | PASS |
| `go test -p=2 -tags badassitron_m3 ./... -count=1` | PASS |
| `go test -p=1 -tags badassitron_legacy,badassitron_m3 ./... -count=1` | PASS (initial multi-command invocation timed out; separate rerun completed) |
| `go vet` default, M3, legacy+M3 | PASS |
| `go test -race -p=2 -tags badassitron_m3 ./... -count=1` | PASS |
| `go test -race -p=2 -tags badassitron_legacy,badassitron_m3 ./... -count=1` | PASS |
| `check-api.py` | PASS — all 66 exported declarations and signatures identical |
| `check-selection.py` | PASS — five tag configurations, no mixed arithmetic |
| External fixture with M3 and legacy+M3 | PASS using temporary offline module config |
| Frozen decimal oracles | 3,518 + 1,725 cases, zero unexpected differences |
| Independent mathematical checks from those oracles | 389 + 1,634, zero arithmetic mismatches (skipped/lossy items explicitly tracked) |
| `check_doc_links.py` | PASS (run after final documentation updates again) |

The temporary validation copies lowered Go version declarations to 1.23 **only outside the distributable**. The release `go.mod` remains `go 1.27` with `github.com/jokruger/dec128 v1.4.0` and no `replace`; `go.sum` retains the official checksums supplied by the user. This checkpoint has **not** been tested on Go 1.27.1 with the downloaded official module. Existing M2-D official logs do not validate M3-D.

## Decisions and next gate

- **D57 — ACCEPT (localized):** operation-specific ordinary unsigned division state mapping and Go value conversions, under the pinned upstream v1.4.0 semantics and guarded by tests.
- **D58 — ACCEPT:** expand benchmark paths to include overflow, zero divisor, and 64-bit division; preserve direct-backend measurements under both builds.
- **D59 — REJECT:** branching `Mul` into Jokruger `Mul64` when `other.Hi == 0` improves exact cases but prevents wrapper inlining and regresses full-width overflow. No speculative path dispatch retained.
- **D60 — NO DEFAULT CUTOVER:** M3 remains opt-in. Run the entire official test/vet/race/client suite and the 11-path comparison under Go 1.27.1 with the published Jokruger v1.4.0 module; review historical code provenance separately before M7 publication. M3 need not block M2 because the default decimal engine already imports upstream primitives directly.

Official-host commands from repository root:

```sh
GOWORK=off bash tools/migration-m3-uint128/verify.sh
BENCHTIME=250ms COUNT=7 bash tools/migration-m3-uint128/bench-three-way.sh reports/migration/m3d-official-three-way
cat reports/migration/m3d-official-three-way/comparison.md
cat reports/migration/m3d-official-three-way/inline-summary.md
```
