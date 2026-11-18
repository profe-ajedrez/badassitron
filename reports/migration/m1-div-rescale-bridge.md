# M1 bridge extension: division and explicit quantization

Date: 2026-11-17. Source snapshot: `badassitron_M1_add_sub_bridge.zip`.

## Scope and source policy

This delivery extends only `internal/dec128bridge`. The public `dec128` package,
`dec256`, `dec512`, and `optimus` are unchanged. The module still declares
`require github.com/jokruger/dec128 v1.4.0`. **No Jokruger source is copied,
vendored, or checked into Badassitron; there is no committed `replace`.**
The supplied `dec128-main.zip` is used only as a local module replacement for
experiments and has not been authenticated as the v1.4.0 tag.

## Added interfaces

- `DivAtPrecision(a, b, defaultPrecision)`: historical truncated division using
  Jokruger's imported `uint128.MulCarry` and `QuoRem256By128`; reattempts with
  canonical operands, preserves public error precedence. Precision is **an
  explicit argument**, because the legacy process-wide `defaultPrecision` has
  no public read API. A direct `DivStrict(a,b)` would silently disagree with
  programs calling `SetDefaultPrecision`; no such misleading API was added.
- `RescaleRoundStrict(value, scale, mode) (result, inexact)`: maps all seven
  historical rounding modes by **name**, not numeric ordinal, and delegates
  to upstream `RescaleRoundInexact`, preserving the explicit `inexact` flag.
  Invalid mode returns legacy `InvalidFormat`; invalid scale returns
  `PrecisionOutOfRange`. No upstream global arithmetic policy is read.

## Measured validation

- `TestDivAtPrecisionDifferential`: 25,144 division pairs, comparing coefficient,
  public sign, stored scale, and error; zero mismatches at the historical
  default precision of 19.
- `TestDivAtPrecisionExactRationalTruncation`: 4,500 finite results checked
  against independent `math/big.Rat` division followed by truncation toward
  zero at the returned scale; zero mismatches.
- `TestRescaleRoundStrictDifferential`: 1,400 boundary cases and 18,000
  deterministic pseudo-random cases covering all seven rounding modes;
  outputs and inexact flags match the historical API.
- `TestRescaleRoundStrictErrors`: verifies the old invalid-mode and
  out-of-range-scale errors.
- Full `go test ./... -count=1`, `go vet ./...`, and
  `go test -race ./internal/dec128bridge -count=1`: passed in the local
  **Go 1.23.2 experimental setup**. This is **not** a Go 1.27 validation.

### Local benchmark, three samples, ns/op (lower is better)

System: Linux amd64, Intel Xeon Platinum 8573C, Go 1.23.2.

| Operation | Bridge samples | Historical samples | Allocations |
| --- | --- | --- | --- |
| Add | 7.254 / 7.800 / 7.101 | 3.413 / 3.595 / 3.577 | 0 |
| Sub | 7.223 / 6.894 / 6.055 | 3.041 / 3.030 / 2.921 | 0 |
| Div | 13.00 / 12.48 / 12.17 | 13.66 / 14.00 / 13.55 | 0 |

Benchmarks are limited to fixed operands and do not characterize all operand
ranges. Raw output is retained in this report's source delivery log.

## Explicit remaining blockers

1. **Signed zero:** historical `Dec128` can store a negative sign bit on a
   zero at nonzero scale, but `IsNegative()` returns false for zero and the
   representation is private. A bridge confined to the public API cannot
   reconstruct that bit. The differential comparisons cover **publicly
   observable state**, not private signed-zero bit identity. Do not claim
   bit-for-bit equivalence or activate the public wrapper until resolved.
2. **Default precision:** `DivAtPrecision` only matches `Div` when the caller
   supplies the active historical default. M2 requires explicit ownership of
   this setting in the compatibility facade.
3. **Dependency validation:** upstream ZIP is an unauthenticated `main`
   snapshot, not a verified v1.4.0 module; the project has no verified
   module checksum in `go.sum`. An official module download and compatible
   toolchain are required before release.
4. **Coverage:** explicit `DivRoundWithStatus`, `MulRoundWithStatus`, parsing,
   SQL/JSON and other decimal methods have not been migrated. Full M1 and M2
   acceptance criteria are not met.
5. **Source attribution:** direct use of Jokruger primitives remains under
   its own upstream module license; the historical-code provenance audit
   remains a separate open item.

## Reproduction (on a machine with the official dependency and Go 1.27)

```sh
GOWORK=off go test ./internal/dec128bridge -count=1
GOWORK=off go test ./... -count=1
GOWORK=off go vet ./...
GOWORK=off go test -race ./internal/dec128bridge -count=1
GOWORK=off go test ./internal/dec128bridge -run '^$' \
  -bench 'Benchmark(DivAtPrecision|DivLegacyForBridge|AddStrict|AddLegacy|SubStrict|SubLegacy)$' \
  -benchmem -count=3
```
