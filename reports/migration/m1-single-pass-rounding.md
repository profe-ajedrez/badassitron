# M1 checkpoint: single-pass rounded multiplication and compatibility surface

Date: 2026-11-17. Source: `badassitron_M1_rounded_bridge.zip`.

## Scope

- Refactored experimental `internal/dec128bridge.MulRoundWithStatus` to calculate a **single** 256-bit product with imported `github.com/jokruger/dec128/uint128.MulCarry`, then divide or multiply the coefficient once with imported `uint128` primitives. It no longer calls upstream `MulRound` after computing the remainder.
- Implemented the seven historical rounding decisions using exact quotient/remainder and preserving the historical `inexact` status, error codes, sign of nonzero results, and stored output scale. This is Badassitron-specific compatibility logic, **not copied from upstream**.
- Added `migration_surface_test.go`: constructors, text/JSON roundtrips, SQL `ScanInto`, finite conversions, error-code conversions, parsing distinction and negative-zero limitation.
- Added `mul_round_independent_test.go`: independent `math/big.Int` product/quantization for all seven rounding modes. 13,979 finite checks; inputs requiring coefficient overflow are excluded from finite-result checks. The preexisting differential test additionally covers overflow, invalid mode and precision, and NaNs.
- No change to public `dec128`, `dec256`, `dec512`, `optimus` arithmetic.

## Validation

| Check | Result |
|---|---|
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| `go test -race ./internal/dec128bridge -count=1` | PASS |
| 30,000 random differential pairs + seeded boundary matrix | PASS |
| 10,000 finite conversion roundtrips | PASS |
| 13,979 independent finite product-round checks | PASS |

## Benchmark (Go 1.23.2, Xeon Platinum 8573C)

Three repeated runs, 250 ms nominal per result, zero allocations per operation:

| Benchmark | Samples ns/op | Median ns/op |
|---|---|---:|
| `MulRoundBridge` single-pass | 16.31, 17.51, 16.89 | 16.89 |
| `MulRoundLegacy` | 19.32, 19.38, 18.94 | 19.32 |
| `DivRoundBridge` | 34.35, 35.37, 34.51 | 34.51 |
| `DivRoundLegacy` | 21.08, 21.53, 21.64 | 21.53 |

Only the indicated benchmark input is measured. Multiplication is now faster than legacy **for this one case**; the division bridge remains slower. No general performance guarantee or universal speed ranking is implied.

## Dependency and toolchain limitations

- The checked-in `go.mod` continues to declare `require github.com/jokruger/dec128 v1.4.0` with no `replace`; upstream sources are **not present in this deliverable**.
- Validation was performed with a **temporary local module replacement** pointing at the separately supplied `dec128-main.zip`. This ZIP is a `main` snapshot; it is not verified byte-for-byte against v1.4.0.
- Go in this execution environment is 1.23.2; the required `go 1.27` in Badassitron was lowered only in temporary test files and restored before packaging. Upstream's `go.mod` was similarly adjusted temporarily. Formal verification with the official tagged module and Go 1.27 is still outstanding.
- `go.sum` has not been populated with validated upstream module hashes; a reproducible public module build still requires normal `go mod download` or `go mod tidy` in a network-enabled environment.

## Remaining blockers before M2

1. `IsNegative()` conceals the sign bit of a stored signed zero. A public-only converter cannot prove roundtrip fidelity of internal signed zero. Decide whether to add an internal export/import path or explicitly normalize it.
2. Parsing, JSON and SQL currently remain handled by historical Badassitron implementations: the tests freeze their behavior but do **not** establish that upstream implements the same external contract.
3. DivRound bridge cost should be investigated before public cutover; do not replace the historical public `dec128` yet.
4. Verify the precise upstream v1.4.0 artifact and Go 1.27 toolchain, and execute the complete migration oracle across two separate module revisions.
