> **Historical migration record (archived at M7-H).** This file contains

> **Archived engineering record.** This document describes a dated source
> checkpoint, not the currently recommended API or release state.
> For supported usage, see the [documentation index](../../../docs/README.md).
> time-specific stage status, proposed code and, in places, old rollback
> tag instructions that **do not apply to the current source tree**.
> The authoritative state is [M7-H release status](../../../docs/RELEASE_STATUS.md);
> do not follow an old "current" or "pending" claim as a live rule.

## M3-D state translation and success/error benchmarks (2026-11-17)

## M3-F current default backend — performance rule

The goal is minimal overhead **relative to a direct call to imported Jokruger**,
not just improvement against Badassitron's former algorithms. This requires
paired three-way benchmarks of identical operands and allocation counts.
Unsigned error codes and public method signatures preclude claiming zero
wrapper overhead. Preliminary Go 1.23.2 short samples are noisy; especially
`Add`, `WideQuoRem`, and `FromString` short input need official Go 1.27.1
retesting. The long parser now delegates directly and avoids per-digit facade
arithmetic. Representative 256/512-bit and Optimus consumer benchmark comparisons
are available in the M3-F report; short local runs had contradictory changes
for division, so representative official measurements are mandatory. See [M3-F report](../m3f-default-imported-uint128-20261117.md).


The [M3-D report](../m3d-division-status-three-way-20261117.md)
records eleven three-way operations, including division-by-zero and overflow.
Local Go 1.23.2 medians show improvement on regular `QuoRem`, but `QuoRemZero`
and `WideOverflow` still have substantial facade overhead. The three-sample
comparison is diagnostic and **does not authorize the M3 default cutover**.
Official Go 1.27.1 benchmarks and provenance review remain open.

## M3-C multiplication wrapper and three-way benchmarks (2026-11-17)

The optional `uint128.Mul` facade wrapper became inlinable under Go 1.23.2
(cost 107 → 77), but exact-success cases and overflow still show >10% sampled
regressions against the historical facade. Measurements used a temporary
upstream reference; official Go 1.27.1 evidence is pending. See the
[M3-C report](../m3c-mul-inline-20261117.md) and
[seven-case comparison](../m3c-local-three-way/comparison.md).

# Migration performance — preliminary baseline (M0) and experimental M1

## M3-A opt-in `uint128` performance (not a release gate pass)

On local Go 1.23.2, the public `uint128` adapter showed +90.3% on Add and
+26.1% on overflowing Mul, despite 0
allocations. These are five-sample median microbenchmarks, not an official
Go 1.27.1 measurement. **Do not switch public `uint128` by default**.
See the [M3-A report](../m3-uint128-opt-in-20261117.md)
and [raw benchmarks](../m3-local-bench/comparison.md).

## M2-E current M6 status (unpublished; 2026-11-17)

The [user-supplied official M2-D comparison](../m2e-supplied-evidence/comparison.md)
used Go 1.27.1 and Jokruger v1.4.0 on AMD Ryzen 7 7735HS. M2-D met the
provisional +10% trigger on 19 of 20 measured operations. The sole exception
was `Rescale/Truncate`: legacy 3.54 ns/op, M2 3.98 ns/op (+12.6%, +0.44 ns/op),
zero heap allocations in both. The experimental direct-uint64 quotient fast path
was rejected after local tests showed no reliable improvement, and it is not
present in M2-E. Switching build tags does not change that algorithm.

Reproduce representative paired benchmarks after promoting the source default:
`bench-m6.sh` now compares `legacy.txt` to `default.txt`. Go 1.27.1/M2-E
rerun and a documented release decision for the residual exception remain open.
See the [M2-E report](../m2e-default-promotion-20261117.md).

**No post-migration comparison has been performed.** Jokruger v1.4.0 is not
installed, so results below measure the *legacy Badassitron* implementation
only. The authoritative raw data is
[`reports/migration/m0-fastpath.txt`](../m0-fastpath.txt).

- Host: Intel Xeon Platinum 8573C, linux/amd64, 5 visible CPUs.
- Compiler: Go 1.23.2, temporary `go.mod` version change to allow local tests.
- Command: `GOMAXPROCS=1 go test ./dec128 -run '^$' -bench '^BenchmarkFastPath$' -benchtime=100ms -count=2 -benchmem`.
- This is a **smoke benchmark**, not a stable statistical reference. Use at
  least five runs at a longer benchtime on a supported Go 1.27 host for M6.
- The results cannot be compared numerically with Ryzen 7 7735HS / Go 1.27.1
  measurements in [`FASTPATH_V4.md`](FASTPATH_V4.md) as a migration regression.

The M6 comparison must include `jokruger/dec128` direct, the legacy snapshot,
and the new Badassitron facade, with the same inputs, exactness/rounding policy,
scales and machine. Do not treat direct upstream default-policy operations as
semantically equivalent to Badassitron strict operations.


## Experimental M1 parser optimization — 2026-11-17

[Checkpoint report and per-shape table](../m1-parser-signed-zero-20261117.md),
with [before raw samples](../m1-parser-before-20261117.txt)
and [after raw samples](../m1-parser-after-20261117.txt).
On the local AMD EPYC 9V74 with Go 1.23.2 and `GOMAXPROCS=1`, the long
fractional input median dropped from **103.20 to 43.71 ns/op** for the bridge,
with **0 B/op and 0 allocs/op**. In the after run the historical median was
48.15 ns/op. The bridge remains somewhat slower on short inputs and an
invalid long-scale rejection path; it is materially faster on sampled large
integers. This is an **experimental** parser comparison, not a post-cutover
public-library performance claim. The dependency was supplied from an
unverified local upstream snapshot with temporary module adjustments.

## M2 isolated candidate / M1 experimental delegation rejected — 2026-11-17

The [M2 candidate checkpoint](../m1-final-m2-candidate-a-20261117.md)
contains a same-host five-sample microbenchmark comparison on AMD EPYC 9V74,
Go 1.23.2, against the historical public package and direct upstream source
snapshot. Medians (ns/op): Add 3.97 historical vs 4.08 candidate; Mul 4.23 vs
4.56; Div 10.48 vs 10.92; Parse 22.89 vs 18.94. All sampled calls report
zero allocations. Upstream direct operations are not always equivalent under
the library's exactness and rounding policy.

An attempted centralization of the M1 bridge through the candidate was fully
**reverted**, since paired bridge benchmarks regressed by 54–138% in this
workload despite functional equivalence. No such delegation appears in the
archive. Full M6 remains open: broader input distributions, source provenance,
Go 1.27, published v1.4.0, independent profiles and full downstream cases.


## M2-B public opt-in performance matrix — 2026-11-17

The [M2-B experimental public cutover report](../m2b-experimental-public-cutover-20261117.md)
contains a **20-case same-host** public-method benchmark matrix in addition to
the earlier isolated-candidate figures. It uses Go 1.23.2,
`GOMAXPROCS=1`, five 80ms samples per build and a separate five-sample
100ms follow-up for optimized `RescaleRound`.

Notable paired medians (legacy → M2, ns/op): mixed-scale Add 10.60 → 11.33,
long parse 79.49 → 44.05, large Mul 22.15 → 25.62, SQL ScanInto 24.77 →
28.52, full-width Sqrt 189.30 → 224.10. The specialized half-even
RescaleRound changed from 12.76 before the final optimization to 5.64 ns/op
in an **unpaired** rerun (legacy earlier median 10.29). Regenerate a fully
paired 1.27/v1.4.0 comparison before promotion.

**M6 remains partial.** The optimized candidate reduces several earlier
large regressions but still exceeds the provisional +10% budget on sampled
wide multiplication, SQL scanning and sqrt. Benchmark outputs are
[default](../m2b-bench-default-20261117.txt),
[M2 before final round optimization](../m2b-bench-tagged-round-before-20261117.txt),
and [specialized rounding after](../m2b-bench-round-after-20261117.txt).


## M2-C: independent sqrt exactness, direct SQL scan and full matrix

See [M2-C report](../m6c-performance-provenance-20261117.md)
for implementation, paired five-sample benchmark and differences. The historical benchmark collector and comparator, now available only in the
separately archived M3-F source ZIP, require matching benchmark names and
recorded environment; they flagged >10% median regressions for review,
**not** automatic acceptance. The current [release benchmark](../../../tools/release/README.md)
compares the facade against imported Jokruger directly. The new code needs an
additional Go 1.27.1 / official v1.4.0 run. No default cutover occurred.


## M2-D: fast division and independent rational verification

See [M2-D checkpoint](../m2d-fast-division-provenance-20261117.md) and [paired 20-case comparison](../m2d-bench-comparison-20261117.md). The M2-D short local run (Go 1.23.2, five 75ms samples/variant) had no >10% *review* flags and reduced the previously flagged `DivRound` regression to +0.8% in the sampled workload. Longer paired and Go 1.27.1 official-v1.4.0 measurements remain mandatory before any default cutover.

## M3-E architectural boundary

The user-supplied official Go 1.27.1 M3-D comparison is archived at
[three-way evidence](../m3e-official-evidence/comparison.md),
with [compiler inlining](../m3e-official-evidence/inline-summary.md).
Nine of eleven measured M3-D facade operations exceeded the provisional 10%
regression threshold. No global upstream performance claim follows from these
operands or their error-path distribution. M3-E keeps the public unsigned
facade historical by default while the M2 decimal coefficient engine uses
Jokruger directly.

A separate internal/public constructor comparison is recorded in the
[M3-E report](../m3e-architecture-resolution-20261117.md).
Four short local Go 1.23.2 samples show no material change in internal
`fromParts` time and a lower median for public `New`. The sample is **not**
Go 1.27.1 release evidence. Use the official M3 three-way and M2 M6 benchmark
scripts after verification.
