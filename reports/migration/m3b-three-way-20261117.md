# M3-B — Three-way benchmark and wrapper inlining audit

## Scope and provenance

Continuation of M3-A ZIP. This experiment compares three separate paths under identical benchmark inputs:

- A: historical Badassitron `uint128` facade (no M3 tag);
- B: imported `github.com/jokruger/dec128/uint128` called directly;
- C: Badassitron M3 facade using the imported backend (`-tags badassitron_m3`).

**Important:** Measurements reported here used **Go 1.23.2** and the supplied upstream source snapshot in a temporary local module replacement. The snapshot has not been established to be byte-identical to the published v1.4.0 tag. No local replacement is included in the delivered `go.mod`, nor is upstream source bundled. The official Go 1.27.1 benchmarks supplied previously concern M3-A, not this optimization.

## Three-way diagnostic — before wrapper change

Five 120 ms samples per path, AMD EPYC 9V74, GOMAXPROCS=1, medians in ns/op. All samples reported 0 B/op and 0 allocs/op.

| Operation | A Historical | B Upstream direct | C M3 adapter | C vs A | C vs B (same tagged build) |
|---|---:|---:|---:|---:|---:|
| Add | 1.315 | 1.218 | 1.823 | +38.6% | +47.1% |
| MulExact (`Mul64`) | 1.833 | 1.761 | 2.233 | +21.8% | +21.8% |
| MulOverflow | 2.521 | 2.468 | 3.643 | +44.5% | +46.1% |
| QuoRem | 10.930 | 9.993 | 11.470 | +4.9% | +9.0% |
| WideQuoRem | 11.490 | 12.330 | 12.500 | +8.8% | +4.2% |

B is measured separately in both builds; the percentages against B use the direct upstream measurement from the **tagged** build, rather than the historical-build measurement in the B column.

## Compiler evidence

With `-gcflags=-m=2`, Go 1.23.2 reports:

- historical `Uint128.Add`: inlinable, cost 40;
- initial M3 `Uint128.Add`: **not** inlinable, cost 90 (>80);
- upstream `Uint128.Add` and `toUp`/`fromUp` conversions: inlinable;
- `Uint128.Mul` and `QuoRem` wrappers: not inlinable (cost above budget).

Therefore treating all the regression as heap conversion overhead would be incorrect. The conversion helpers inline and all five benchmarks allocate zero bytes. **Wrapper inlining and error/status mapping are plausible material causes**, although disassembly and performance counter evidence would be necessary to allocate exact shares of cost.

## Small M3-only code change

`Add`, `Add64`, `Sub`, `Sub64`, and `Mul64` now convert directly between the two public-limb value types, return early on non-OK upstream status, and keep historical error codes. They still delegate the arithmetic to Jokruger. The `-m=2` report now marks all five wrappers inlinable; `Add` has cost 60.

No change to the default uint128 implementation, public signatures, or M2 decimal backend. Other M3 methods remain as in M3-A.

## After wrapper change

Five comparable local 120 ms samples, medians:

| Operation | A historical (earlier control) | C optimized M3 | Delta | Direct upstream, tagged build |
|---|---:|---:|---:|---:|
| Add | 1.315 | 1.585 | +20.5% | 1.298 |
| MulExact (`Mul64`) | 1.833 | 2.126 | +16.0% | 2.073 |
| MulOverflow (unchanged wrapper) | 2.521 | 3.854 | +52.9% | 2.542 |
| QuoRem (unchanged wrapper) | 10.930 | 11.490 | +5.1% | 10.940 |
| WideQuoRem (unchanged wrapper) | 11.490 | 13.410 | +16.7% | 12.870 |

These are **not** controlled enough to attribute every percentage change to code: the control A samples were collected in a preceding run and the independent B timings also vary. Nevertheless, the reduction in Add wrapper inlining cost is compiler-confirmed. No promotion of M3 is authorized by these numbers.

## Validation performed in this environment

- `go test ./... -count=1` — PASS.
- `go test -tags badassitron_m3 ./... -count=1` — PASS.
- `go test -p 2 -tags badassitron_legacy,badassitron_m3 ./... -count=1` — PASS (an earlier unconstrained attempt timed out).
- `go vet -tags badassitron_m3 ./...` — PASS.
- `go test -tags badassitron_m3 ./dec128/uint128 -count=1` — PASS.

These runs used a temporary local module configuration with Go 1.23.2, not the official published dependency; they must be repeated with Go 1.27.1 and the official module.

## Run in the official environment

From the repository root, without any local `replace` directives:

```bash
GOWORK=off bash tools/migration-m3-uint128/verify.sh
BENCHTIME=250ms COUNT=7 bash tools/migration-m3-uint128/bench-three-way.sh reports/migration/m3b-official-three-way
```

Inspect `comparison.md`, `inline-historical.txt`, `inline-adapter.txt`, and `environment.txt`. Keep the original M3-A benchmark output for historical comparisons; do not use a single percentage to generalize backend speed.

## Decision

Maintain M2 as the default decimal engine and M3 as opt-in. The direct backend is generally faster than the M3 adapter in these measured operations. The adapter's inlining is a concrete actionable issue, but M3 `Mul` overflow still requires profiling and focused optimization before any cutover decision.
