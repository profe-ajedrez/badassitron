# Badassitron

Badassitron is a Go library for decimal arithmetic and calculation workflows.
Its decimal packages keep base 10 values without using binary floating point
for decimal arithmetic. The repository also contains older calculation
packages built around decimal and floating point values.

Start with `optimus` for adaptive decimal arithmetic. Use the fixed width
packages when a fixed coefficient size is part of your design.

## Requirements

This snapshot declares Go 1.27 in `go.mod`. 

```sh
go get github.com/profe-ajedrez/badassitron
```

## Choose a decimal type

| Package | Coefficient capacity | Maximum decimal scale | Main use |
| --- | --- | --- | --- |
| `dec128` | 128 bits | 19 | Existing fixed width decimal code |
| `dec256` | 256 bits | 77 | Fixed width decimal values with more capacity |
| `dec512` | 512 bits | 154 | Large fixed width decimal values |
| `optimus` | Adaptive 64, 128, 256, 512 bits | 154 | General decimal work with automatic promotion |

A decimal value is a nonnegative integer coefficient, a sign, and a decimal
scale. For example, a coefficient of 12345 and a scale of 2 represent 123.45.
`Precision()` returns the stored decimal scale, not the number of significant
digits.

## Quick start with Optimus

```go
package main

import (
    "fmt"

    "github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
    value := optimus.FromString("12.345")
    fmt.Println(value.RoundBank(2))

    price := optimus.FromString("98")
    rate := optimus.FromString("1.18")
    quantity := optimus.FromInt(2000)
    net := price.Div(rate).Mul(quantity).
        Mul(optimus.FromString("0.9")).RoundBank(2)
    fmt.Println(net)
}
```

Output:

```text
12.34
149491.53
```

`FromString` returns an `Optimus` value. Invalid input is represented by
`NaN`, so check `IsNaN()` before using untrusted input. Use `NewFromString`
when you want a Go error instead.

```go
value, err := optimus.NewFromString("123.4500")
if err != nil {
    panic(err)
}
fmt.Println(value.String())
fmt.Println(value.StringFixed())
```

Output:

```text
123.45
123.4500
```

## Exactness and rounding

`Optimus` keeps an exact result when it is representable. If ordinary
arithmetic needs an approximation, it rounds half even at the highest scale
that fits. If the integer magnitude cannot fit, the result is `NaN`.
A division with a repeating decimal uses the working precision, which
starts at 19. Exact terminating divisions can preserve more digits.

Choose explicit operations when rounding is part of a business rule or a
published output format.

```go
value := optimus.FromString("2.5")
fmt.Println(value.Round(0))
fmt.Println(value.RoundBank(0))

result, inexact := optimus.FromString("1").DivRoundWithStatus(
    optimus.FromString("3"), 2, optimus.RoundingHalfEven,
)
fmt.Println(result, inexact)
```

Output:

```text
3
2
0.33 true
```

`Round` uses half away from zero. `RoundBank` uses half even. `Rescale` and
`Trunc` use truncation toward zero. A `WithStatus` boolean reports discarded
nonzero information. It is not an error value. Check `IsNaN()` separately.

Use `SetDefaultPrecision` only when a process wide setting is appropriate.
For calculations that must behave consistently across concurrent work, use
explicit precision operations instead.

## SQL values

Use `optimus.SQLDecimal` when exact database input matters. It distinguishes
SQL NULL from numeric zero and rejects floating point scan input.

```go
var amount optimus.SQLDecimal
err := amount.Scan("123.4500")
if err != nil {
    panic(err)
}
if amount.Valid {
    fmt.Println(amount.Decimal.StringFixed())
}
```

The older `Optimus.Scan` maps NULL to zero and accepts floating point input.
Its behavior is retained for compatibility. SQL correctness also depends on
the driver and database column returning or storing the full decimal text.

## Decimal conversions

Use `FromDec128`, `FromDec256`, and `FromDec512` to create `Optimus` values.
Use `Dec128Exact`, `Dec256Exact`, and `Dec512Exact` to narrow without silent
rounding. A conversion that cannot preserve the value returns an error.

## Examples

See [the examples directory](examples/README.md) for complete programs with
plain English explanations, run commands, and expected output. They cover
Optimus, rounding, invoice calculations, JSON, SQL, Dec128, Dec256, Dec512,
exact conversions, concurrency, and the existing decimal calculation pipeline.

## Documentation

- [Optimus user guide](docs/OPTIMUS_STAGE.md). Precision, limits, NaN, SQL,
  concurrency, and migration guidance.
- [GoDoc reference](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus).
  Public methods and examples.
- [Existing calculation workflow](docs/CALCULATION_PIPELINE.md).
  Staged handlers, required order, and known gaps.
- [Wide decimal design](docs/DECIMAL_WIDE_STAGE.md).
- [Unsigned integer design](docs/WIDE_INTEGER_FOUNDATION.md).
- [Dec128 precision](dec128/PRECISION.md).
- [Dec128 SQL scanning](dec128/SQL_SCANNING.md).
- [Optimus concurrency](optimus/CONCURRENCY.md).
- [Changelog](CHANGELOG.md).
- [Optimus v8 audit](OPTIMUS_V8_PHASE_C_PRECISION_AUDIT.md).

Older audit files record how the code changed. They are historical reports,
not the primary reference for current behavior.

## Verification

```sh
go test ./optimus/... ./dec128/... ./dec256/... ./dec512/... ./withdec128/...
go test -race ./optimus/... ./dec128/... ./dec256/... ./dec512/...
go vet ./optimus/... ./dec128/... ./dec256/... ./dec512/...
```

The Optimus v8 audit passed the related test, race, and vet checks with an
external Go 1.23 compatibility modfile. Native Go 1.24 verification was not
available for that audit. The original Optimus v8 audit recorded existing
compile errors in `withfloat64/internal/math.go`. Those errors were subsequently
corrected without changing the legacy package API.

## Scope and compatibility

`optimus`, `dec128`, `dec256`, and `dec512` provide numeric building blocks.
They do not encode tax rules, money rules, or a fixed rounding policy for a
particular country. Select the required operation and rounding point in your
application.

Optimus v8 retains its public API. The v8 correction fixes overflow
validation in scale aligned quotient and remainder calculations. See the
audit and changelog for the exact case, test evidence, and performance cost.

## License

See [LICENSE](LICENSE).

## Experimental Dec1024 backend

An additive 1024-bit decimal backend is available under [`dec1024`](dec1024/README.md).
It supports a 308-place fractional scale via `uint16`, a fixed-width
`uint1024` coefficient, and lossless widening from Dec128/256/512.
Existing backends and Optimus remain unchanged.
