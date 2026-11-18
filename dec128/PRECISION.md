# Dec128: precision and rounding

`dec128.Dec128` stores a signed decimal with a **128-bit coefficient** and
at most **19 fractional decimal places**. Use it for fixed-width
integrations; use [Optimus](../docs/OPTIMUS.md) if adaptive coefficient
capacity is preferable.

For a working program, run:

```sh
go run ./examples/06_dec128
```

[Read the code and expected output](../examples/06_dec128/README.md).

## Parse and calculate

```go
price, err := dec128.NewFromString("98.00")
if err != nil {
    log.Fatal(err)
}
mode := dec128.RoundingHalfEven
unit := price.DivRound(dec128.FromString("1.18"), 19, mode)
subtotal := unit.MulRound(dec128.FromInt(2000), 19, mode)
net := subtotal.MulRound(dec128.FromString("0.90"), 19, mode)
published, inexact := net.RescaleRoundWithStatus(2, mode)
if published.IsNaN() {
    log.Fatal(published.ErrorDetails())
}
fmt.Println(published.StringFixed(), inexact) // 149491.53 true
```

This snippet assumes imports for `fmt`, `log`, and `dec128` as in the
[complete program](../examples/06_dec128/main.go).

The example keeps more precision during intermediate calculations than in
the final printed amount. The choice of a 19-place working scale, a 2-place
published scale, and half-even rounding is **an example policy**, not a
jurisdiction-specific financial rule.

## Which method should I use?

| Need | API | Note |
| --- | --- | --- |
| Parse untrusted text | `NewFromString` | Returns a Go error for invalid text |
| Round multiplication | `MulRound(other, scale, mode)` | Explicit output scale |
| Round division | `DivRound(other, scale, mode)` | Explicit output scale |
| Round square root | `SqrtRound(scale, mode)` | Explicit output scale |
| Round an existing decimal | `RescaleRound(scale, mode)` | Can add trailing zeros or round discarded digits |
| Check whether nonzero digits were lost | `*WithStatus` variants | Returns `(value, inexact)`; check `IsNaN()` too |
| Truncate an existing value | `Rescale(scale)` or `Trunc(scale)` | Do not assume these round half-even |

All explicit-scale operations still have a 128-bit coefficient limit.
A result that cannot be represented can become `NaN`; the `inexact` flag
is **not** an error indicator. The ordinary `Div` and `Sqrt` methods have
legacy truncating behavior and consult a process-wide default scale, initially
19; explicit rounding methods do not rely on that default.

## Ties and stored scale

```go
value := dec128.FromString("1.245")
fmt.Println(value.RescaleRound(2, dec128.RoundingHalfEven).StringFixed()) // 1.24
fmt.Println(value.RescaleRound(2, dec128.RoundingHalfAwayFromZero).StringFixed()) // 1.25
```

`Precision()` returns the stored **fractional scale**, not significant
digits. `StringFixed()` displays that scale. Adding zeros using
`RescaleRound` may fail when the coefficient is already at capacity.
For modes and errors see [Rounding and errors](../docs/ROUNDING_AND_ERRORS.md).

## Quotients, remainders, and SQL

`QuoRem` returns an integer quotient truncated toward zero and a remainder
with the dividend's sign. An unrepresentable integer quotient can make
`QuoRem` return invalid values even when the exact remainder is representable;
`Mod` can still return that remainder.

The existing `Dec128.Scan` has a **value receiver** and therefore cannot
update the caller's `Dec128` through `rows.Scan(&value)`. Use
`dec128.ScanDestination(&value)` or `dec128.ScanInto(&value, src)` instead.
See [SQL scanning](SQL_SCANNING.md) and the [SQL program](../examples/05_sql/README.md).

For a broader comparison with `dec256`, `dec512`, and `optimus`, see
[Choosing a decimal type](../docs/DECIMAL_TYPES.md).
