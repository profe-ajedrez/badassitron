# Fixed-width unsigned foundation: Uint256 and Uint512

This is a historical foundation stage document. Dec256, Dec512, and Optimus
are now implemented. Use the README and the Optimus user guide for current
usage and limitations.

This stage adds **integer primitives only**. `Dec256`, `Dec512`, and `Optimus`
were not yet implemented at this stage. It did **not** change the representation,
API, precision, rounding, or semantics of the existing `dec128` package.

## Locations and public value types

- `dec128/uint128.Uint128` (preexisting): `Lo uint64`, `Hi uint64`.
- `dec256/uint256.Uint256`: `Lo uint128.Uint128`, `Hi uint128.Uint128`.
- `dec512/uint512.Uint512`: `Lo uint256.Uint256`, `Hi uint256.Uint256`.

All use little-endian halves, are fixed-size comparable value types and have
zero allocation for arithmetic in the measured benchmarks. Their zero values
are valid integers equal to zero.

## Arithmetic contracts

- `Add`, `Sub`, `Mul`: return `(value, errors.Error)`; overflow/underflow is
  **reported, never silently wrapped**. Error results are the zero value.
- `MulCarry`: return `(low, high)` of the **full 512-bit or 1024-bit product**;
  the pair represents `low + (high << width)`.
- `Div`, `Mod`, `QuoRem` and their `*64` variants: exact unsigned integer
  arithmetic; division by zero uses `errors.DivisionByZero`. `u=q*v+r` and
  `0<=r<v` hold when the divisor is nonzero.
- `Lsh`/`Rsh`: truncate to the fixed width; shifts >= width return zero.
- Bitwise operations, rotates, reversal and bit counts mirror `uint128`.
- `FromString` accepts base-10 ASCII digits, including leading zeros and an
  empty string (zero), as existing `uint128.FromString` does. Invalid input
  and overflow are errors, not truncations.
- `BigInt`/`FromBigInt` are conversion boundaries. The newer Uint256 and
  Uint512 converters do not mutate input. The inherited Uint128 converter
  can mutate its big integer argument, as noted by the Optimus v8 audit.
- Binary serialization is available in both byte orders. Text marshaling and
  `fmt.Scanner` are supported.
- `FromUint128` / `Uint128` (checked) and `FromUint256` / `Uint256` (checked)
  provide explicit widening and narrowing.

The error codes come from `dec128/errors`, preserving the project's existing
error vocabulary. Integer arithmetic itself uses `math/bits`, **not math/big**.

## Decimal capacity and future policy

- `uint256`: 256 bits, maximum coefficient `2^256-1` (78 digits), powers of
  ten `10^0` through `10^77` inclusive.
- `uint512`: 512 bits, maximum coefficient `2^512-1` (155 digits), powers of
  ten `10^0` through `10^154` inclusive.

The agreed **154 fractional positions** and automatic promotion/nearest
rounding apply to the later `Optimus` decimal API, not to these unsigned
integer primitives. A coefficient's storage width and a decimal's fractional
scale must be treated as independent constraints.

## Implementation and performance limits

`internal/uintwide` supplies fixed-size unsigned limb arithmetic, full-width
schoolbook products and exact long division. Division uses a single-limb fast
path; general multi-limb division currently uses a shift/subtract algorithm.
It is correct against a big-integer oracle but **has not yet been optimized**
with Knuth/Burnikel-Ziegler division. This matters for the next stages.

`Uint256` and `Uint512` deliberately have matching method names but not
identical Go types to `Uint128` (the `Lo`/`Hi` field types become wider). This
is not a drop-in source replacement for a typed `Uint128` value.

## Original stage goals

1. Define new `Dec256` and `Dec512` public types with the **same everyday API
   conventions** as `Dec128`, using their matching fixed-width coefficients.
2. Explicitly define precision/scale range, sign, signed zero, `NaN`, overflow
   and rounding semantics without changing the historical `Dec128` API.
3. Build `Optimus`, prioritizing the smallest sufficient representation,
   promoting without loss and performing banker's rounding when an exact
   finite decimal result cannot be represented within the 154-scale/512-bit
   bounds. Preserve the `Dec128`-style API for ordinary operations.
4. Benchmark small-number fast paths and multi-limb division before selecting
   the implementation strategy for large decimal workloads.

## Verification

With the Go version declared in `go.mod` (currently Go 1.27):

```sh
go test ./dec128/... ./dec256/... ./dec512/... ./withdec128/...
go test -race ./dec128/... ./dec256/... ./dec512/... ./withdec128/...
go vet ./dec128/... ./dec256/... ./dec512/... ./withdec128/...
go test ./dec256/... ./dec512/... -bench '^BenchmarkWide' -benchmem
```

The `withfloat64` package is outside the scope of this historical stage.
Its earlier compiler errors were subsequently corrected.
