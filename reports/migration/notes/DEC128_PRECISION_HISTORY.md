# Explicit-precision arithmetic in `dec128`

> **Archived implementation notes.** Use the [documentation index](../../../docs/README.md)
> for current behavior and runnable examples.

This guide covers explicit-precision operations in the current `dec128`
package. These operations were introduced in earlier stages and their public
contracts are retained by the current imported-Jokruger-backed facade. For a brief working example
start with the [Dec128 example](../../../examples/06_dec128/README.md). The older
`Dec128` constructors, methods, rounding functions, JSON behavior, and
`Mul` / `Div` contracts remain available. The coefficient still uses `uint128.Uint128`, and `MaxPrecision`
remains 19 fractional digits.

## New operations

```go
import "github.com/profe-ajedrez/badassitron/dec128"

mode := dec128.RoundingHalfEven

unitValue := dec128.FromString("98.00").DivRound(dec128.FromString("1.18"), 19, mode)
gross := unitValue.MulRound(dec128.FromString("2000"), 19, mode)
net := gross.MulRound(dec128.FromString("0.90"), 2, mode)

fmt.Println(net.StringFixed()) // 149491.53
```

The methods are:

- `MulRound(other Dec128, prec uint8, mode RoundingMode) Dec128`
- `DivRound(other Dec128, prec uint8, mode RoundingMode) Dec128`
- `MulRoundWithStatus(other Dec128, prec uint8, mode RoundingMode) (Dec128, bool)`
- `DivRoundWithStatus(other Dec128, prec uint8, mode RoundingMode) (Dec128, bool)`

`WithStatus` returns `inexact=true` when the mathematically exact intermediate
has a nonzero remainder at the requested scale. This is independent of whether
the rounded coefficient was increased. Errors continue to be represented by
`NaN`; callers must check `IsNaN()` and must not interpret `inexact` as an error
indicator.

Supported modes:

| Mode | Meaning |
| --- | --- |
| `RoundingTowardZero` | Truncate magnitude |
| `RoundingAwayFromZero` | Increase magnitude for every nonzero discarded fraction |
| `RoundingFloor` | Toward negative infinity |
| `RoundingCeil` | Toward positive infinity |
| `RoundingHalfTowardZero` | Nearest, with ties toward zero |
| `RoundingHalfAwayFromZero` | Nearest, with ties away from zero |
| `RoundingHalfEven` | Nearest, with ties to even |

Both operations compute with a complete 256-bit intermediate, quantize once,
and check whether the final *coefficient* fits into `uint128`. No binary
floating point, `math/big` arithmetic, heap allocations, or process-wide
precision settings are required by these operations. Their results still have
at most 19 fractional digits and a 128-bit coefficient. This does not grant
arbitrary-precision intermediate chains: each returned `Dec128` is a new,
explicit quantization boundary.

## Compatibility and safety fixes

- `Compare` now compares full scaled magnitudes (up to 256 bits) before
  applying the sign, so large negative operands compare correctly even when
  rescaling a coefficient to a common exponent would overflow 128 bits.
- `Equal` and `Compare` treat positive and negative representations of zero
  as the same numerical value, including zeros with different scales.
- `SetDefaultPrecision` now uses an atomic store, and legacy `Div` and `Sqrt`
  use atomic loads. This prevents data races, but a mutable global setting is
  still sensitive to concurrent timing. Prefer explicit-precision operations
  when independent calculations require different policies.

## Non-goals and known limitations

- Existing `Mul`, `Div`, `Rescale`, `Round`, string formatting, JSON and SQL
  methods were not replaced with the new rounding policy.
- Existing `Rescale` can discard fractional digits. Use a deliberate rounding
  operation before rescaling if such truncation is not intended.
- The existing `Scan` method has a **value receiver**, so it cannot assign
  parsed data to its caller's variable. Changing the receiver to `*Dec128`
  changes Go method sets and therefore requires a separate compatibility
  decision; this update leaves the method intact.
- SUNAT examples are test vectors for precision behavior, not tax algorithms
  inside `dec128`. Tax rates, taxable bases and calculation order belong to
  the calling application.

## Tests and benchmarks

New tests cover full-width coefficient comparisons against `math/big.Rat`,
all rounding modes, ties and negative values, exact/inexact statuses,
unsupported modes, precision errors, overflow, and the `149491.53` SUNAT
example. Randomized multiplication and division are checked against
`math/big.Rat.FloatString` using half-away-from-zero rounding. Concurrency is
checked with Go's race detector. All tests use an external reference only in
the test package.

Run with the Go version declared in the current `go.mod` (Go 1.27):

```sh
go test ./dec128/... ./withdec128/...
go test -race ./dec128/... ./withdec128/...
go test ./dec128/tests/unit -run '^$' -bench 'Benchmark(MulRound|DivRound|MulLegacy|DivLegacy)$' -benchmem
```

## Second iteration: square root and integer division safety

The legacy `Sqrt()` continues to use `SetDefaultPrecision` and **truncates**
its square root result to that precision. It now uses a complete 256-bit
radicand, an overflow-safe integer Newton iteration, and handles coefficients
as large as `uint128.Max` without looping indefinitely. This is a correctness
fix, not a change of the rounding policy of `Sqrt()`.

`QuoRem()` continues to use the quotient truncated toward zero, with a
remainder of the dividend's sign. When the divisor's magnitude exceeds the
dividend's magnitude, it now returns `(0, dividend)` directly; scaling that
divisor to a common exponent could previously overflow despite an exactly
representable result. `Mod()` shares this correction.

The historical `Dec128.Scan` value receiver remains unmodified to preserve
the public method set. For functional SQL scanning see [SQL_SCANNING.md](../../../dec128/SQL_SCANNING.md)
and use `dec128.ScanDestination(&amount)` or `dec128.ScanInto(&amount, source)`.

The added independent property tests compare decimal arithmetic to
`math/big.Rat` and square roots to `math/big.Int.Sqrt`; `math/big` is not
used in production `Dec128` arithmetic.

## Third iteration: explicit square-root rounding

The existing `Sqrt()` continues to **truncate** at the process-wide default
precision. Its signature, early-return behavior for zero and one, and its
rounding policy have not changed.

For a calculation with an explicit quantization boundary, use:

```go
x := dec128.FromString("2.25")
result, inexact := x.SqrtRoundWithStatus(0, dec128.RoundingHalfEven)
fmt.Println(result, inexact) // 2 true: sqrt(2.25) = 1.5, a tie to even

rounded := dec128.FromString("2").SqrtRound(19, dec128.RoundingHalfEven)
fmt.Println(rounded) // 1.4142135623730950488
```

Additive methods:

- `SqrtRound(prec uint8, mode RoundingMode) Dec128`
- `SqrtRoundWithStatus(prec uint8, mode RoundingMode) (Dec128, bool)`

Both accept precisions from 0 through `MaxPrecision`. The `inexact` flag is
true when the exact square root cannot be represented at the requested scale,
**even when the rounded coefficient remains unchanged**. The flag is not an
error code; check `IsNaN()` for errors. All seven existing rounding modes are
supported. Since a nonnegative square root is returned, floor and toward-zero
have the same result, as do ceil and away-from-zero.

Rounding decisions use the full 256-bit radicand and exact integer comparisons,
including real half ties such as `sqrt(2.25) = 1.5` and `sqrt(6.25) = 2.5`.
They do not use `float64`, `math/big`, the global precision, or heap
allocations. A 128-bit coefficient limit still applies to results.

## Modulo with an unrepresentable quotient

`QuoRem` requires a 128-bit integer quotient and returns two `NaN` values when
that quotient cannot fit. `Mod` does not require the quotient to be returned.
It now calculates its exact remainder independently, including when the
integer quotient would need more than 128 bits:

```go
max := dec128.FromString("340282366920938463463374607431768211455")
tiny := dec128.FromString("0.0000000000000000003")
q, r := max.QuoRem(tiny) // NaN, NaN: quotient is too large
remainder := max.Mod(tiny) // valid, exact decimal remainder
```

The remainder's sign follows the dividend, as before. `Mod` continues to
return `NaN` for a zero divisor or NaN inputs. No existing method signatures
have changed.

## Fourth iteration: explicit rescaling and inexact reporting

The original `Rescale(prec)` continues to **truncate toward zero** when reducing
fractional digits and to return `NaN` on coefficient overflow when increasing
fractional digits. Its signature and legacy behavior are unchanged.

New, additive methods allow callers to specify rounding **at the rescale
boundary** without changing process-wide settings:

```go
x := dec128.FromString("1.245")
rounded, inexact := x.RescaleRoundWithStatus(2, dec128.RoundingHalfEven)
fmt.Println(rounded.StringFixed(), inexact) // 1.24 true

// 1.245 is an exact half tie; away-from-zero produces 1.25 instead.
y := x.RescaleRound(2, dec128.RoundingHalfAwayFromZero)
fmt.Println(y.StringFixed()) // 1.25
```

- `RescaleRound(prec uint8, mode RoundingMode) Dec128`
- `RescaleRoundWithStatus(prec uint8, mode RoundingMode) (Dec128, bool)`

Both accept `prec` in the inclusive range `0..MaxPrecision` and all seven
`RoundingMode` values already used by `MulRound`, `DivRound`, and `SqrtRound`.
Increasing the scale is **exact** and cannot produce `inexact=true`, but can
return `NaN` if multiplying the coefficient by a power of ten would overflow
128 bits. Keeping the same scale is exact. Reducing the scale divides the
coefficient by `10^(inputScale-prec)` and rounds the resulting quotient once,
using the *exact* remainder. An exact zero is rescaled with its internal sign
bit preserved, as in the legacy method.

`inexact=true` means that the original value was not exactly representable at
the requested scale, including when the rounded result is zero or the rounded
coefficient did not increase. It does **not** indicate an error; use `IsNaN()`
for invalid precision, invalid mode, pre-existing NaN or coefficient overflow.
Errors return `inexact=false` for this operation. Never use the boolean alone
to determine success.

The implementation uses `uint128`/`uint64` coefficient and remainder arithmetic
and delegates its rounding decision to the same internal exact quantizer used
by multiplication and division. It does not invoke `math/big` at runtime,
`float64`, or the global precision setting. The return value remains a 128-bit
coefficient and can have at most 19 fractional decimal places. The operation
is a deliberate quantization boundary; it does not recover precision already
lost in an earlier calculation.

### Suggested financial calculation pattern

```go
mode := dec128.RoundingHalfEven
unit := dec128.FromString("98.00").DivRound(dec128.FromString("1.18"), 19, mode)
subtotal := unit.MulRound(dec128.FromString("2000"), 19, mode)
net := subtotal.MulRound(dec128.FromString("0.90"), 19, mode)
published, inexact := net.RescaleRoundWithStatus(2, mode)
fmt.Println(published.StringFixed(), inexact) // 149491.53 true
```

The example is a *decimal precision regression*, not an embedded SUNAT tax
rule. Jurisdiction-specific policies still belong to the caller.
