# Getting started in ten minutes

Badassitron works with **base-10 decimals**, avoiding accidental binary
floating-point rounding in decimal arithmetic. This guide uses `optimus`,
which automatically chooses a coefficient size. You need Go 1.27 or newer.

## 1. Run a working program

In your Go module:

```sh
go get github.com/profe-ajedrez/badassitron
```

Create `main.go`:

```go
package main

import (
    "fmt"
    "log"
    "strings"

    "github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
    input := "19.95"
    if strings.TrimSpace(input) == "" {
        log.Fatal("price is required")
    }
    price, err := optimus.NewFromString(input)
    if err != nil {
        log.Fatal(err)
    }
    subtotal := price.Mul(optimus.FromInt(3))
    if subtotal.IsNaN() {
        log.Fatal(subtotal.ErrorDetails())
    }
    fmt.Println(subtotal.StringFixed())
}
```

```sh
go run .
```

Output: `59.85`. The parser's empty-string-to-zero behavior is retained for
compatibility, so explicitly reject empty input when your application needs
a required value.

The complete repository also contains a [basics program](../examples/01_optimus_basics/README.md)
that demonstrates parsing, formatting, errors, and automatic widening.

## 2. Decide where to round

A decimal stores an integer coefficient and a *scale*. `19.95` has coefficient
`1995` and scale `2`. `Precision()` reports the stored scale, not the count of
significant digits.

```go
x := optimus.FromString("2.5")
fmt.Println(x.Round(0))     // 3  (ties away from zero)
fmt.Println(x.RoundBank(0)) // 2  (ties to even)
```

For a result that **must have exactly two stored decimal places**, use:

```go
printed := optimus.FromString("12.5").RescaleRound(2, optimus.RoundingHalfEven)
fmt.Println(printed.StringFixed()) // 12.50
```

`RoundBank(2)` does not pad an input with a smaller stored scale; `RescaleRound`
does. The latter can fail with `NaN` if a new coefficient does not fit.
For an invoice or financial report, decide whether to round **per line** or
**at the total**, then apply that policy explicitly. Compare with the
[runnable invoice example](../examples/03_invoice/README.md).

## 3. Know what errors look like

`NewFromString` returns `(value, error)` for parsing. Most arithmetic methods
instead return a decimal and encode an error as `NaN`.

```go
value := optimus.FromInt(1).Div(optimus.FromInt(0))
if value.IsNaN() {
    fmt.Println(value.ErrorDetails()) // division by zero
}
```

For `MulRoundWithStatus`, `DivRoundWithStatus`, `RescaleRoundWithStatus`, and
similar methods, the boolean means **inexact**: nonzero digits were dropped.
It is not a success/error flag.

```go
third, inexact := optimus.FromInt(1).DivRoundWithStatus(
    optimus.FromInt(3), 2, optimus.RoundingHalfEven,
)
if third.IsNaN() {
    log.Fatal(third.ErrorDetails())
}
fmt.Println(third.String(), inexact) // 0.33 true
```

An integer that cannot fit even in 512 bits may produce `NaN` rather than
wrapping around. Adaptive storage does not mean unlimited precision.

## 4. Keep SQL and JSON exact

For `database/sql`, use `optimus.SQLDecimal` when `NULL` must be distinct from
zero and decimal digits must be preserved:

```go
var price optimus.SQLDecimal
err := price.Scan("123.4500")
if err != nil {
    log.Fatal(err)
}
fmt.Println(price.Valid, price.Decimal.StringFixed()) // true 123.4500
```

`SQLDecimal.Scan` rejects float inputs; require a driver that returns exact
decimal text or integers and use an appropriate `DECIMAL`/`NUMERIC` column.
For JSON, follow the [JSON example](../examples/04_json/README.md) and do
not decode money or decimal strings through `float64`.

The older `Dec128.Scan` has a value receiver. If you use `Dec128`, pass
`dec128.ScanDestination(&value)` to `rows.Scan`, rather than `&value`.
The [SQL example](../examples/05_sql/README.md) demonstrates both cases.

## 5. Choose a narrower type only when needed

- `optimus`: adaptive, up to a 512-bit coefficient and scale 154. Start here.
- `dec128`: fixed 128-bit coefficient, scale at most 19.
- `dec256`: fixed 256-bit coefficient, scale at most 77.
- `dec512`: fixed 512-bit coefficient, scale at most 154.

Conversions between packages are **not** implicit. Use exact conversion
methods such as `Dec128Exact()` when loss is not permitted; see
[the conversion example](../examples/09_conversions/README.md).

For precise arithmetic differences, read [Decimal behavior](DECIMAL_BEHAVIOR.md).
For every runnable example, use [the example index](../examples/README.md).
For method signatures, see [GoDoc](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus).
