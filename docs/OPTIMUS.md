# Optimus: adaptive decimal arithmetic

Use `optimus` when you do not need to choose a fixed coefficient width.
Start with [Getting started](GETTING_STARTED.md), or run
`go run ./examples/01_optimus_basics`.

## Read values safely

```go
amount, err := optimus.NewFromString("123.4500")
if err != nil {
    log.Fatal(err)
}
fmt.Println(amount.String())      // 123.45
fmt.Println(amount.StringFixed()) // 123.4500
```

This snippet assumes imports for `fmt`, `log`, and `optimus` as in the
[complete program](../examples/01_optimus_basics/main.go). Use
`NewFromString` for external input. `FromString` is convenient for trusted
literals but reports malformed input as `NaN`. Historically, an empty string
is interpreted as zero; reject blank required input yourself.

Values use a coefficient and a decimal *scale*. `Precision()` returns the
scale (four for `123.4500`), **not** the number of significant digits.
`String()` removes trailing fractional zeros; `StringFixed()` retains them.

## Do arithmetic and check the result

```go
subtotal := optimus.FromString("19.95").Mul(optimus.FromInt(3))
if subtotal.IsNaN() {
    log.Fatal(subtotal.ErrorDetails())
}
fmt.Println(subtotal.StringFixed()) // 59.85
```

Arithmetic returns new values. Optimus selects a 64-, 128-, 256-, or
512-bit coefficient representation as needed, up to scale 154. Width is a
storage detail, not a promise of unlimited precision. Ordinary operations
may round when exact representation cannot fit; some results become `NaN`.

## Choose rounding deliberately

```go
value := optimus.FromString("12.5")
fmt.Println(value.Round(0))     // 13 (half away from zero)
fmt.Println(value.RoundBank(0)) // 12 (half even)

fixed := value.RescaleRound(2, optimus.RoundingHalfEven)
fmt.Println(fixed.StringFixed()) // 12.50
```

`RoundBank(2)` does not add fractional zeros to values with a smaller stored
scale. `RescaleRound(2, mode)` can do so, but may return `NaN` if the new
coefficient cannot fit. `DivRound`, `MulRound`, and `SqrtRound` allow explicit
scale and mode; `WithStatus` variants also report discarded nonzero digits.
Always check `IsNaN()` independently of the inexact flag.

[Run rounding examples](../examples/02_rounding/README.md) ·
[Compare line and total rounding](../examples/13_rounding_boundaries/README.md)

## Save, transfer, and convert

For JSON, use the [JSON example](../examples/04_json/README.md) and avoid
intermediate `float64` decoding when original decimal digits matter.
For nullable SQL values, prefer `optimus.SQLDecimal`, which distinguishes
`NULL` from zero and rejects floating-point scans.

`optimus.FromDec128`, `FromDec256`, and `FromDec512` import fixed-width values.
`Dec128Exact`, `Dec256Exact`, and `Dec512Exact` return an error when a smaller
format cannot represent the value exactly. See the
[conversion example](../examples/09_conversions/README.md).

## Shared settings and limitations

The process-wide default working precision initially uses 19 decimal places
for selected operations. `SetDefaultPrecision` is atomic, but concurrent
changes can affect separate operations. Prefer explicit scales for
reproducible work. [Concurrency notes](../optimus/CONCURRENCY.md) explain
safe sharing.

Ordinary `Div` and `Sqrt` use different rounding policies from some
fixed-width operations. JSON `null` maps to numeric zero in the inherited
`Optimus` contract. A `NaN` cannot be round-tripped as a normal decimal
through text or JSON. For exact edge cases, see
[Decimal behavior](DECIMAL_BEHAVIOR.md) and [Rounding and errors](ROUNDING_AND_ERRORS.md).

[GoDoc and API](GODOC_AND_EXAMPLES.md) · [All examples](../examples/README.md)
