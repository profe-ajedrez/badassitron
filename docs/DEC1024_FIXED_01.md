# BADASSITRON-DEC1024-FIXED-01 — Experimental fixed-width core

Status: **implemented and tested as an experiment, not approved as a production replacement**.

## Goal and scope

Eliminate runtime `math/big` and the dependency on `internal/decimalwide` from the Dec1024 decimal algorithms without changing public signatures, coefficient capacity, supported scale, rounding policies, or the behavior of other backends. The preexisting Dec1024 ZIP is the immutable baseline.

## Design

- Keep `Dec1024` as a comparable 136-byte value with an inline 1024-bit coefficient and a `uint16` scale (0–308).
- `fixedNum` is an internal, stack-eligible `[64]uint64` (4096-bit) temporary, not a new public `Dec4096` API. No `unsafe` and no global mutable working buffers are introduced.
- `internal/uintwide.QuoRem` is unchanged. New `QuoRemWide` implements normalized multi-limb division (Knuth's algorithm D) for 4096-bit internal values.
- Multiplication uses exact products and the number of significant limbs where appropriate; scale adjustment uses `10^19` blocks to avoid hundreds of one-digit divisions.
- `fixedFitExact` removes trailing zeros only when necessary for exact representation. `fixedQuantize` performs a single final rounding using the remainder and denominator, preserving the `inexact` contract.
- Integer square root uses monotone Newton iteration on a fixed-width integer quotient, with an exact midpoint comparison for rounding (no float64 approximations).
- Public `Uint1024.FromBigInt` and `Uint1024.BigInt` remain available: these are **explicit conversion APIs**, not calls made by Dec1024's calculation path. `internal/decimalwide` remains unchanged for Dec256/Dec512.
- Existing `Mul`/`MulRound` fixed-width fast paths were retained after measuring regressions from an attempted universal variable-width multiplication.

### Width bounds

Given coefficients of at most 1024 bits and scales of at most 308:

- An exact coefficient product occupies at most 2048 bits.
- Scale alignment for addition/modulo requires at most roughly 2048 bits (plus one possible carry).
- The scaled numerator of division and of square root uses at most `1024 + ceil(616 * log2(10))`, i.e. **3071 bits**.
- Exact square-root rounding comparison requires a small additional margin; 4096 bits accommodate it.
- Every intermediate exponent and shift is bounded by the public scale contract. Exponentiation by squaring checks each multiplication for representability; it does not retain arbitrary intermediate powers.

These bounds justify an internal 4096-bit temporary. They do *not* imply that every `fixedNum` remains physically on the stack under every future Go compiler: the compiler decides that through escape analysis.

## Actual changes

- Changed: `dec1024/dec1024.go` (uses fixed implementations in calculation paths and removes its `math/big` import, including for narrowing and Int64 conversion).
- Added: `dec1024/fixed_arith.go`, `dec1024/fixed_differential_test.go`, `dec1024/fixed_wide_bench_test.go`.
- Added: `internal/uintwide/divwide.go`, `internal/uintwide/divwide_test.go`.
- Unchanged: Optimus, Dec128, Dec256, Dec512 and the existing `internal/uintwide/arith.go` implementation.
- No new dependencies or changes to exported API signatures.

## Verification

Execution environment: Go **1.23.2**, linux/amd64, Intel Xeon Platinum 8573C. The original sources declare `go 1.27`; verification used *temporary copies only* with a local `go 1.23` directive. The deliverable retains `go 1.27` and has **not** been verified using Go 1.27.

- `go test ./...`: PASS.
- `go vet ./...`: PASS.
- `go test -race ./dec1024 ./internal/uintwide`: PASS.
- Selected `uint1024` race tests: PASS; this is not a claim of a full repository `-race` run.
- Differential tests against the old `internal/decimalwide` engine: PASS (large random operands, seven rounding modes, scale boundaries, operations, error and `inexact` checks).
- New `QuoRemWide` limb-division oracle tests: 10,000 randomized comparisons against `math/big`, PASS.
- Zero-heap assertions using `testing.AllocsPerRun` for ordinary and wide calculation scenarios: PASS.

See the evidence ZIP for raw logs, benchmark data and CPU profile.

## Benchmark summary (medians; ns/op)

| Scenario | Original | Fixed | Original allocs/op | Fixed allocs/op | Interpretation |
|---|---:|---:|---:|---:|---|
| Basic Add | 193.6 | 240.1 | 0 | 0 | Regression in a small input |
| Basic Mul | 991.6 | 1001.0 | 0 | 0 | Nearly unchanged |
| Basic MulRound | 1490.0 | 1616.0 | 0 | 0 | Moderate regression |
| Basic DivRound | 1747.0 | 1756.0 | 0 | 0 | Nearly unchanged |
| Wide Compare | 1068 | 452.6 | 14 | 0 | Faster |
| Fitable wide Add | 1741 | 1392 | 18 | 0 | Faster |
| Exact 2000-bit Mul | 103820 | 13699 | 597 | 0 | Substantially faster |
| Rounded wide Mul | 3891 | 4753 | 25 | 0 | Regression |
| 2048-bit scaled Div | 4218 | 3534 | 21 | 0 | Faster |
| Very-wide Mod | 1992 | 2581 | 23 | 0 | Regression |
| Wide Sqrt | 12666 | 17671 | 35 | 0 | Regression, improved from initial implementation |
| PowInt(40) | 2325 | 6333 | 41 | 0 | Significant regression |

Benchmark samples: five on original, three on final variant; `-benchtime=220ms` and `-benchtime=140ms`, respectively, all on the same host. Values are exploratory and subject to scheduling noise. Source CSV includes `B/op`, `allocs/op`, sample counts and percent deltas. These numbers do not prove an overall performance gain in financial applications.

The final square-root CPU profile attributes the largest sampled fraction to `uintwide.QuoRemWide` (approximately 36% flat), followed by fixed decimal scaling. This identifies follow-up optimization opportunities.

## Decision and limits

**Keep as experimental and do not replace production Dec1024 by default yet.** The goal of eliminating `math/big` from the decimal calculation path is achieved in the tested prototype. Allocation-free execution is confirmed only for the tested operations and inputs, not for every public API (formatting and serialization inherently build output buffers). CPU results are mixed, with substantial regressions in exponentiation and some rounding cases.

Recommended next iterations:

1. Investigate a specialized small-coefficient multiplication for `PowInt` that preserves the existing fast path for general `Mul`.
2. Optimize wide rounded multiplication and remainder paths; retain the existing reference as oracle.
3. Expand differential/property tests, especially extreme signed zeros, error ordering and larger exponents.
4. Re-run the full `-race` suite and all benchmarks under Go 1.27 once that toolchain is available, and evaluate actual Iacon/EUCLIDS workloads before adoption.

The experiment does not introduce public `Dec2048` or `Dec4096`; neither is required to hold the internal intermediates.
