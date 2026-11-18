# M1 checkpoint: direct imported-primitive rounded division

Date: 2026-11-17. Starting artifact: `badassitron_M1_single_pass_rounding.zip`.

## Scope

- Replaced the experimental `internal/dec128bridge.DivRoundWithStatus`'s `jk.Dec128.DivRoundInexact` round-trip with direct arithmetic using **imported** `github.com/jokruger/dec128/uint128` primitives: `MulCarry`, `QuoRem256By128`, `QuoRem`, and `Add64`.
- Aligned numerator/denominator scales without lossy intermediate decimal conversion. When a scaled denominator exceeds 128 bits, the integral quotient is zero but its full 256-bit width is retained for correct half-rounding.
- Implemented all seven historical rounding decisions and the `inexact` flag from the exact quotient/remainder. Invalid mode, invalid precision, division by zero, NaN propagation and overflow retain the historical public error semantics.
- Removed the now-unused conversion-based rounding-error helper. The public `dec128` implementation remains entirely unchanged.
- Added an independent `math/big.Rat` verification across the seven rounding modes, including randomly selected large 128-bit coefficients.
- Added a test proving that a signed zero in the historical representation is not losslessly carried through the upstream constructor.

**No Jokruger code was copied into this repository.** Go imports its public package as a module dependency. This bridge still contains Badassitron-owned adapter and rounding-policy glue.

## Verification

| Check | Observed result |
|---|---|
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| `go test -race ./internal/dec128bridge -count=1` | PASS |
| Existing `TestRoundedOpsDifferential` (seeded boundaries + 30,000 random pairs) | PASS |
| New independent `math/big.Rat` division/rounding checks | **55,531 finite results checked, 0 discrepancies** |
| Stored negative zero representation preservation by `ToUpstream` | **NOT POSSIBLE** using upstream `jk.New`; acknowledged and tested |

The independent verification skips 128-bit-overflow NaNs: those are covered by the differential tests, not considered independent arithmetic successes. The 55,531 checked results are individual rounding-mode outcomes, not distinct input pairs.

## Microbenchmark

Intel Xeon Platinum 8573C, linux/amd64, Go 1.23.2, 5 repeated measurements of the same existing benchmark inputs, zero allocations for both implementations:

| Benchmark | Samples (ns/op) | Median (ns/op) |
|---|---|---:|
| Direct-primitives `DivRoundBridge` | 21.52, 21.56, 21.39, 22.20, 21.67 | **21.56** |
| Historical `DivRoundLegacy` | 20.27, 20.13, 20.86, 20.03, 19.97 | **20.13** |

The previous M1 single-pass checkpoint recorded `DivRoundBridge` at 34.51 ns/op. The new implementation is approximately **37.5% faster** than that previous measurement, but is approximately **7.1% slower** than historical Badassitron in this repeated case. Those older and current measurements come from the same general execution environment but are **not** a statistically controlled full workload study. No general speed ranking is claimed.

Raw files: `m1-direct-division-bench.txt`, `m1-direct-division-tests.txt`, `m1-direct-division-vet.txt`, `m1-direct-division-race.txt`.

## Representation decision still required

The historical `legacy.New(zero, scale > 0, true)` stores a negative sign bit although `IsNegative()` returns `false` and public textual output is zero. Jokruger's `jk.New(zero, scale, true)` normalizes the sign to positive. Consequently, `FromUpstream(ToUpstream(value))` cannot promise binary/field-level identity for historical negative zero. The *numeric* and *public formatting* contracts are currently maintained. Before M2 choose one of:

1. **Explicitly normalize zeros** and approve representation-only incompatibility; or
2. Have the new Badassitron wrapper carry its own signed-zero metadata (with constructors/operations guaranteeing updates), independently of upstream.

Do not use reflection/unsafe to inspect upstream private fields. No decision has yet been made to change the public API.

## Dependency and toolchain constraints

The checked-in `go.mod` remains `go 1.27` and `require github.com/jokruger/dec128 v1.4.0`; **no replace directive** is present in the deliverable. During testing a temporary `replace` pointed to the separately extracted `dec128-main.zip`, and the Go directives were temporarily lowered for the locally available Go 1.23.2 compiler. The original declarations are restored before packaging; the upstream source archive is *not* included. This snapshot of upstream `main` has not been verified as byte-for-byte identical to the published v1.4.0 tag. A valid `go.sum` and formal test with the tagged module/Go 1.27 remain outstanding.

## Next gates

- Approve the signed-zero representation policy for the eventual M2 wrapper.
- Obtain and verify the official `v1.4.0` module and checksum in a Go 1.27 environment.
- Finish parser, serialization, SQL and construction adapters without replacing public API methods piecemeal.
- Execute the full historical cross-process oracle and a statistically meaningful performance run before the atomic M2 cutover.
