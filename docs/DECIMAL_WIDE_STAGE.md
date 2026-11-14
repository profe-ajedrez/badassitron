# Wide decimal stage: Dec256 and Dec512

This document records the fixed width decimal stage. Optimus is now
implemented. Use the README and the Optimus user guide for the current
adaptive decimal API.

This stage implements the two new signed fixed-width decimal value types:

- `dec256.Dec256`: coefficient `uint256.Uint256` (256 bits); `MaxPrecision == 77`.
- `dec512.Dec512`: coefficient `uint512.Uint512` (512 bits); `MaxPrecision == 154`.

Both follow the public naming and *ordinary use* of `dec128.Dec128` without
modifying `dec128`, `uint128`, `uint256`, or `uint512`.

## Representation and default precision

A value is `(-1)^neg * coefficient * 10^(-scale)`. The **scale** and the
coefficient width are independent limits. It is valid to store a coefficient
of `1` at scale 154 in `Dec512`. `MaxPrecision` is the maximum *fractional
scale*, not the number of significant digits of the coefficient.

Both new packages start with `SetDefaultPrecision(19)`; their `SetDefaultPrecision`
accepts up to 77 or 154 respectively. The global default affects `Div`,
`Sqrt`, and negative integer powers, **not** `DivRound`, `MulRound`,
`SqrtRound` or `RescaleRound`. The setter uses atomic storage.

## API example

```go
package main

import (
    "fmt"
    "github.com/profe-ajedrez/badassitron/dec256"
    "github.com/profe-ajedrez/badassitron/dec512"
)

func main() {
    a := dec256.FromString("98")
    b := dec256.FromString("1.18")
    fmt.Println(a.DivRound(b, 28, dec256.RoundingHalfEven))

    large := dec512.FromDec256(a)
    fmt.Println(large.Mul(dec512.FromInt(2000)))
}
```

`FromString`, `FromInt`, `FromInt64`, `New`, `NaN`, comparisons, conversions,
`Add`, `Sub`, `Mul`, `Div`, `Mod`, `QuoRem`, `PowInt`, `Sqrt`, `Round*`, `Trunc`,
`Rescale*`, `MulRound*`, `DivRound*`, `SqrtRound*`, `Sum`, `Avg`, `Min`, `Max`,
JSON/text marshaling, `Value`, and SQL `Scan` follow the established names.
The two packages share `dec128.RoundingMode` via an alias.

`Dec512` supports `FromDec256` and `Dec256Exact`; both types support
`FromDec128` and `Dec128Exact`. Narrowing is **exact or error**; trailing
zeros can be removed for exact narrowing. `Scan` takes a pointer receiver
in the new types (and thus works directly with `database/sql.Scanner`);
the historical `Dec128.Scan` is unchanged. `ScanDestination`/`ScanInto` are
also available by the same names as in v4.

## Arithmetic contract

`Add`, `Sub` and `Mul` prefer exact results; they never round implicitly to
avoid overflow. Exact trailing powers of ten can be removed when needed to
fit. `Div` truncates at the process-wide default precision (or higher input
scale, as in `Dec128`). `Sqrt` truncates analogously. `Mod` is exact even when
the integer quotient of `QuoRem` overflows. An unrepresentable result becomes
`NaN(error code)`.

`*Round(..., precision, mode)` methods round once to an *explicit* target
scale. `*WithStatus` adds a boolean indicating a discarded nonzero fraction,
not an error indicator. `RoundBank` implements nearest ties-to-even.

**Important:** automatic nearest rounding on representation overflow is an
implemented behavior of **Optimus**, not a behavior of these fixed-width types.

## Performance and known limits

- The common paths for `Add`, `Sub`, `Mul`, `Compare`, `RescaleRound`, and
  `DivRound` use allocation-free fixed-width coefficient arithmetic.
- `internal/decimalwide` uses `math/big.Int` only when exact intermediate
  magnitudes exceed the fixed-width fast path or for complex operations
  such as square root. This initial fallback prioritizes correctness and is
  not yet an optimized all-limb implementation.
- `uintwide.QuoRem` is currently a shift/subtract long division for general
  multi-limb divisors. Division performance can be improved independently.
- `PowInt` with large negative exponents may return overflow before computing
  a tiny reciprocal. This mirrors a checked-power limitation and needs an
  explicit policy in the later Optimus design.
- The new types are concrete, distinct Go types. They are API-similar to
  `Dec128`, **not** drop-in interchangeable with it at compile time.
- `NaN` preserves errors within an individual decimal type. Widening from an
  existing NaN maps the error to `NotANumber`, since the original error code
  is not a public field.

## Tests

The test suite checks random arithmetic against `math/big.Rat` and integer
square roots, large full-width coefficient cases, all seven rounding modes,
serialization, SQL scanning, conversion, overflow, extrema, the SUNAT
149491.53 case, and concurrent precision settings. Fuzz targets are included.

Run on Go 1.24+:

```sh
go test ./dec128/... ./dec256/... ./dec512/... ./withdec128/...
go test -race ./dec128/... ./dec256/... ./dec512/... ./withdec128/...
go vet ./dec128/... ./dec256/... ./dec512/... ./withdec128/...
go test ./dec256 ./dec512 -run '^$' -bench 'Benchmark(Add|Mul|DivRound)$' -benchmem
```

The original fixed width stage did not include `optimus`. The existing unrelated `withfloat64`
compilation errors remain outside scope.
