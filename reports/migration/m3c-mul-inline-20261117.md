# M3-C — Jokruger-backed `uint128.Mul` facade inlining and expanded three-way evidence

Date: 2026-11-17. Source authority: accumulated M3-B archive. This is an **unpublished experimental checkpoint**.

## Scope

Keep `dec128` M2-E as default. Keep the historical public `dec128/uint128` implementation as default. With `-tags badassitron_m3`, use Jokruger imported primitives and preserve Badassitron's nominal `Uint128` type and all historical error codes. Do not copy upstream source into the repository.

## Change

`dec128/uint128/m3_adapter.go`: simplify the `Mul` wrapper from two `toUp` conversions, one `fromUp` conversion and a success-first return to direct *value type conversions* and an overflow-first return. The actual multiplication still calls `ju.Uint128.Mul` imported from Jokruger. The Go compiler may now inline the **wrapper**, while the upstream multiplication itself remains a normal call in the measured toolchain.

No changes to `MulCarry`, `Mul64`, other arithmetic algorithms, public signatures, or the default unsigned implementation. No `unsafe` or upstream source copy.

Go 1.23.2 compiler `-m=2` observations (cost budget 80):

| Method | Historical | Imported Jokruger | M3-B wrapper | M3-C wrapper |
|---|---|---|---|---|
| `Mul` | cannot inline, cost 91 | cannot inline, cost 95 | cannot inline, cost 107 | **can inline, cost 77** |
| `Add` | can inline, cost 40 | can inline, cost 40 | can inline, cost 60 | can inline, cost 60 |

This demonstrates the wrapper transformation, **not** the percentage of cycles attributable to conversions or a guarantee under Go 1.27.1. Diagnostic source: `m3c-local-three-way/inline-summary.md` and the three raw compiler logs.

## Benchmarks (local development environment, not official)

Go 1.23.2, Linux/amd64, AMD EPYC 9V74, supplied upstream ZIP as a **temporary external module reference**; `GOMAXPROCS=1`. Seven operations, five 150ms repetitions for each mode, medians in ns/op. All measured operations reported 0 B/op and 0 allocs/op.

| Operation | Historical facade | Direct imported Jokruger (M3 build) | M3-C facade | M3-C vs historical |
|---|---:|---:|---:|---:|
| `Add` | 1.221 | 1.218 | 1.541 | +26.2% |
| `MulExact` (`Mul64`) | 1.890 | 1.874 | 2.052 | +8.6% |
| `MulNarrowExact` (`Mul`) | 2.787 | 3.171 | 3.639 | +30.6% |
| `MulHighExact` (`Mul`) | 3.197 | 3.189 | 3.532 | +10.5% |
| `MulOverflow` (`Mul`) | 2.523 | 2.533 | 3.035 | +20.3% |
| `QuoRem` | 10.890 | 10.600 | 11.350 | +4.2% |
| `WideQuoRem` | 12.670 | 12.620 | 12.920 | +2.0% |

A separate **focused** 150ms measurement taken directly before the wrapper change yielded M3-B `MulOverflow` median 3.742 ns/op (five samples). The first focused M3-C run yielded 2.758 ns/op (seven samples): a promising reduction but **not** a stabilized effect size. The subsequent seven-operation matrix recorded 3.035 ns/op; do not substitute the more favorable number for the full comparison. See the raw data in the report directory.

**Decision:** The wrapper is structurally simpler and compiler-confirmed inlinable, but 3/7 sampled operations exceed the provisional +10% threshold (notably `MulNarrowExact`). M3 **must remain opt-in**. The historical facade and the imported backend run through different generated machine code; comparing them requires official Go 1.27.1 profiling and repeated same-host measurements.

## Validation (local snapshot only)

- Complete `go test ./... -count=1` (default) — PASS.
- Complete `go test -tags badassitron_m3 ./... -count=1` — PASS.
- Complete `go test -p 1 -tags badassitron_legacy,badassitron_m3 ./... -count=1` — PASS. An earlier combined run timed out; rerunning with reduced package parallelism completed.
- `go vet -tags badassitron_m3 ./...` — PASS.
- `go test -race -p=2 -tags badassitron_m3 ./... -count=1` — PASS.
- `go test -race -p=1 -tags badassitron_legacy,badassitron_m3 ./...` — **INCOMPLETE (execution limit)**. The completed package results plus separate `optimus`/`withdec128`/`withfloat64` race run passed; this is not represented as a globally completed command.
- API declaration checker — **66/66 signatures identical**; source-selection checker — PASS for all five supported tag combinations.
- External client fixture — PASS with M3 and `legacy+M3`, using a temporary module configuration.
- M3 `Mul` test uses 6,000 deliberate exact/overflow input pairs and independent `math/big`; PASS for both builds. The pre-existing randomized unsigned and wide-division tests also pass.
- Frozen differential oracles invoked under M3: primary **3,518/3,518**, complementary **1,725/1,725**, **0 unexpected differences**; independent checks 389 primary, 1,634 complementary, zero mismatches. Skipped/lossy cases remain labeled in raw outputs.

## Reproducibility and upstream verification limit

This evaluation did **not** use Go 1.27.1, nor did it fetch the published `github.com/jokruger/dec128@v1.4.0` module. The upstream reference ZIP declares Go 1.24 and the repository Go 1.27; both module declarations were lowered **only in the temporary test copy** for local Go 1.23.2 execution. The distributable `go.mod` is restored to `go 1.27`, official `require github.com/jokruger/dec128 v1.4.0`, no `replace`, original official `go.sum`. Neither the ZIP snapshot nor other external module contents are shipped in the deliverable.

Run the following on a host with Go 1.27.1 and access to the published dependency:

```bash
GOWORK=off bash tools/migration-m3-uint128/verify.sh
BENCHTIME=250ms COUNT=7 bash tools/migration-m3-uint128/bench-three-way.sh reports/migration/m3c-official-three-way
```

Inspect `comparison.md`, `inline-summary.md`, `inline-upstream.txt` and `environment.txt`. These commands do not authorize an M3 default cutover or an M7 release. Check provenance and M2-E publication gates separately.

## Decisions

- **D53 (KEEP):** Build-time experimental delegation of `Mul` to imported Jokruger, preserving historical zero-on-overflow and error codes.
- **D54 (OPTIMIZE):** Inline-friendly direct Go value conversion in the wrapper; do not duplicate Jokruger's multiplication algorithm.
- **D55 (EXPAND):** Add successful `Uint128.Mul` and overflow benchmarks alongside the existing `Mul64` case; record direct upstream and facade baselines.
- **D56 (NO PROMOTION):** Complete official Go 1.27.1 tagged-module checks, targeted performance profiling, and licensing/provenance review before changing default `uint128`.
