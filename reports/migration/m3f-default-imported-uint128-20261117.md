# M3-F — Default imported Jokruger backend for public Uint128

Date: 2026-11-17. Source: cumulative M3-E ZIP, not GitHub history. **Unreleased source checkpoint.**

## Objective and decision

Use **one imported backend** for normal 128-bit decimal and unsigned arithmetic,
while keeping the public Badassitron API as close to its historical behavior
as possible. Performance must be assessed against **Jokruger called directly**,
not just against the old Badassitron algorithm.

M3-F **promotes the existing M3 adapter into the default build**. The
historical unsigned implementation is no longer in the normal build. It is
retained temporarily under explicit rollback tags as an independent test
oracle until Go 1.27.1 and release acceptance are complete. No upstream
Jokruger source is vendored, copied or mirrored. The pinned import is
`github.com/jokruger/dec128 v1.4.0`.

The `Uint128` public shape remains `struct{ Lo, Hi uint64 }`; its nominal
Go identity and the numeric values and types of `errors.Error` are unchanged.
`ju.Uint128(u)` and `Uint128(upstreamResult)` are ordinary Go conversions
between matching underlying structs, not allocations or `unsafe` casts.
They do not guarantee zero runtime overhead when a wrapper must translate
upstream statuses or when imported functions cannot be inlined.

## Build selection

| Go build tags | `dec128` | `dec128/uint128` |
| --- | --- | --- |
| none | Jokruger-backed M2 | **Jokruger-backed M3-F** |
| `badassitron_m3` | Jokruger-backed M2 | Jokruger-backed M3-F (compatibility alias) |
| `badassitron_uint128_legacy` | Jokruger-backed M2 | Historic unsigned **rollback only** |
| `badassitron_legacy` | Historic decimal rollback | Historic unsigned rollback |
| `badassitron_legacy,badassitron_m3` | Historic decimal rollback | Jokruger-backed M3-F |

The build-selector and AST tools enforce no mixed arithmetic sources and
exactly the same **66 exported unsigned declarations, including signatures**.
This is **not** a claim of binary compatibility or a release tag.

## Implemented changes

1. Inverted the build tags on the *whole* unsigned arithmetic set
   (`arithmetic.go`, `bitwise.go`, `tools.go`, `m3_adapter.go`). Normal operations
   dispatch to imported Jokruger; no partial historical arithmetic fallback.
2. Moved the historical `FromString` and `StringToBuf` algorithms behind
   the rollback tag. The default public implementations delegate directly
   to imported Jokruger, translating only historical error codes.
3. Preserved historical `FromBigInt`, `Scan`, the public type, constants,
   `Value`-style error semantics, and other compatibility surfaces where
   upstream behavior has not been proven identical. This does **not** preserve
   an old *arithmetic engine* in normal production compilation.
4. Added early checks for ordinary zero-divisor error and definitely overflowing
   256/128 division. These checks preserve the exact upstream/historical
   error precedence while retaining upstream arithmetic for valid operands.
   Error-path performance is not extrapolated to valid divisions.
5. Updated build-source, API and architecture guards; refreshed the official
   verification, external-client, three-way, codec benchmark and CI workflows.
   Historical records M0–M3-E are retained as historical evidence.
6. Added independent codec comparisons (strings and bytes) with malformed,
   overflowing, and zero-padded decimal inputs and wide-error precedence cases.

## Measured correctness (local only)

Local host: **Go 1.23.2, Linux/amd64, Intel Xeon Platinum 8573C**. The
validation was executed in a *separate* throwaway copy of this repository,
with `go 1.23` set locally and `replace` pointing at the supplied upstream
reference ZIP. This snapshot has **not** been proven byte-identical to the
published v1.4.0 tag. The release repository retains `go 1.27`, original
`go.sum`, and no `replace` or `vendor` directory.

| Check | Result |
| --- | --- |
| `go test ./... -count=1`, default imported engines | PASS |
| `go test -tags badassitron_uint128_legacy ./... -count=1` | PASS |
| `go test -tags badassitron_legacy,badassitron_m3 ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| M3 architectural AST/source guard | PASS in source-copy test against unmodified release tree |
| Unsigned API inventory and build selector | PASS, 66 exported declarations |
| Existing unsigned integer tests and new codec/precedence tests | PASS |
| Main frozen oracle | 3,518 / 3,518; **0 unexpected differences** |
| Extended frozen oracle | 1,725 / 1,725; **0 unexpected differences** |
| Independent decimal exact checks | 389 + 1,634 checks; **0 mismatches** (skips/lossy cases recorded by oracle) |
| Full global `go test -race -p=2 ./...` | **INCOMPLETE**: environment time limit; no test failure reported in completed packages |
| `-race` for `dec128` bridge/candidate, unsigned, `dec256`, `dec512` | PASS by package groups |
| `-race` for `optimus`, adapters and architecture tests | PASS by package groups |
| Official Go 1.27.1 published module on M3-F source | **NOT RUN HERE** |

The codec parity test includes 35,000 deterministic random input strings,
72,000 additional valid/zero-padded and numeric inputs, fixed boundary
strings and byte-input variants. It also compares formatting with upstream.
This is a differential test against the **supplied** upstream source; the
historical-tag run checks the pre-migration paths against the same reference.

## Performance: direct Jokruger is the target

Local arithmetic matrix: three short samples of 120 ms, `GOMAXPROCS=1`,
Go 1.23.2, measured in separate tagged builds. **These are diagnostic, not
release medians.** In particular, changes in the direct-upstream control
between builds show material run-to-run variation; do not interpret small
cross-build differences as causal.

[Three-way comparison](m3f-local-evidence/three-way-comparison.md) and
[unsigned codec comparison](m3f-local-evidence/codecs-comparison.md) retain
raw ns/op and allocation data. A few observed cases:

| Case | Historical facade | Imported M3-F facade | Direct Jokruger in M3-F build |
| --- | ---: | ---: | ---: |
| `Add` | 1.491 ns | 2.753 ns | 1.717 ns |
| `MulOverflow` | 3.381 ns | 3.685 ns | 3.393 ns |
| `QuoRem` | 16.690 ns | 17.320 ns | 18.100 ns |
| `QuoRemZero` | 2.432 ns | 2.630 ns | 2.575 ns |
| `WideOverflow` | 3.997 ns | 2.913 ns | 4.165 ns |

There are **still overheads** for simple `Add`, certain multiplication and
normal wide division inputs. The migration does not claim universal parity.
A nominal distinct type returning historical error codes cannot be an exact
alias of the upstream type. Keep examining inlining and optimizing only where
measurements with Go 1.27.1 substantiate an opportunity.

For codecs, short same-host samples suggested long `FromString` improved from
about 106 ns to 71 ns per operation, close to Jokruger's direct call, with
zero allocations; **short parsing was too noisy to conclude**. The 19-digit
chunk formatting delegates directly but remains subject to shape and
compiler effects. Do not generalize from one string length or CPU.

### Wider consumer performance

[Representative `dec256`, `dec512`, `optimus` comparison](m3f-local-evidence/consumer-comparison.md)
uses identical M2 decimal engines and alternates only the `uint128`
implementation. Short local three-sample measurements were mixed, including
possible regressions in `optimus` division. A second alternating short run
showed that the same `optimus` case reversed advantage from one run to the
next, demonstrating **significant host noise**; no consumer performance pass
or failure is claimed. The official `bench-consumers.sh` script records a
longer, paired Go 1.27.1 comparison for these end-to-end workloads.

## Outstanding blockers

- Rerun **M3-F code itself** with the published Jokruger v1.4.0 and Go 1.27.1,
  including full race, external-module client, frozen oracles and official benchmarks.
- Review any persistent >10% performance deltas (especially `Add` and
  wide-division success), in context of representative decimal/Optimus use.
- Confirm original Badassitron code provenance, notices and permission to
  redistribute/retire the historical rollback. Go module checksum
  verification alone is not proof of authorship.
- After all release gates pass, move the historical unsigned and decimal
  rollback sources outside the production module into a separately preserved
  reference, or remove them once the frozen oracle suffices. Do **not** drop
  rollback tests before the official verification is recorded.

## Reproduce on the official Go 1.27.1 host

From the root of this delivered repository:

```bash
GOWORK=off bash tools/migration-m2-cutover/verify.sh
GOWORK=off bash tools/migration-m3-uint128/verify.sh
BENCHTIME=250ms COUNT=7 bash tools/migration-m3-uint128/bench-three-way.sh reports/migration/m3f-official-three-way
BENCHTIME=250ms COUNT=7 bash tools/migration-m3-uint128/bench-codecs.sh reports/migration/m3f-official-codecs
BENCHTIME=250ms COUNT=7 bash tools/migration-m3-uint128/bench-consumers.sh reports/migration/m3f-official-consumers
# Optional and slower compiler diagnostics:
INLINING=1 BENCHTIME=250ms COUNT=7 bash tools/migration-m3-uint128/bench-three-way.sh reports/migration/m3f-official-inlining
```

Review `comparison.md` in both outputs. No publication/version tag was made.

## M0–M7 status

| Stage | State after M3-F |
| --- | --- |
| M0 — Contract/provenance | Partially validated; historical rights not fully established |
| M1 — Differential bridge | Existing oracles and bridge pass local validation |
| M2 — Public `dec128` | Jokruger-backed **default**, official final rerun pending |
| M3 — Public `uint128` | **Jokruger-backed default in code**; historical rollback tagged; official final rerun pending |
| M4/M5 — Wide decimals and `optimus` | Not rewritten; consumer tests pass with imported unsigned backend |
| M6 — Performance | Three-way and codec diagnostics completed locally; official M3-F review pending |
| M7 — Release | BLOCKED pending official validation and provenance |

## Decisions

D65: promote complete imported public unsigned arithmetic source as default;
D66: keep nominal API and context-specific errors rather than claiming
zero-cost wrappers; D67: delegate string conversion codecs and preserve
historic golden compatibility; D68: require official host/performance/provenance
gates before release. See [MIGRATION_DECISIONS](notes/MIGRATION_DECISIONS.md).
