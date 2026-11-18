# Decimal recipes

Every command below runs from the repository root and has a complete source
file and expected output. These recipes choose example policies; your
application must decide which policies are appropriate.

## 1. Read an amount from text

Use `NewFromString`, not `FromString`, when the caller needs a Go parsing
error. Check for blank input separately if zero and missing are different.

```go
text := "19.95"
amount, err := optimus.NewFromString(text)
if err != nil {
    log.Fatal(err)
}
fmt.Println(amount.StringFixed()) // 19.95
```

**Run:** `go run ./examples/01_optimus_basics` — [code](../examples/01_optimus_basics/main.go).

## 2. Multiply a price by a quantity

```go
price := optimus.FromString("12.50")
total := price.Mul(optimus.FromInt(3)).
    RescaleRound(2, optimus.RoundingHalfEven)
if total.IsNaN() {
    log.Fatal(total.ErrorDetails())
}
fmt.Println(total.StringFixed()) // 37.50
```

**See:** [the quick start](GETTING_STARTED.md) or
[the basic arithmetic program](../examples/01_optimus_basics/main.go).

## 3. Round with an explicit rule

A half-even tie at `2.5` rounds to `2`; the same tie with `Round(0)` rounds
to `3`. To *store* exactly two fractional places, call `RescaleRound`.

**Run:** `go run ./examples/02_rounding` — [explanation](../examples/02_rounding/README.md).

## 4. Decide whether to round lines or totals

Three exact values `1.005`, `2.005`, and `3.005` produce **6.00** if each
line is rounded half-even to two places first, versus **6.02** if the exact
sum is rounded once. The library cannot choose the business rule for you.

**Run:** `go run ./examples/13_rounding_boundaries` —
[code and output](../examples/13_rounding_boundaries/README.md).

## 5. Calculate a discount and tax

The example takes three units at 19.95, a 10% discount, and an illustrative
18% tax. It rounds the discount and tax separately to two places (half-even).
The final total is **63.57**. The program is **not** tax-law guidance.

**Run:** `go run ./examples/03_invoice` — [steps and output](../examples/03_invoice/README.md).

## 6. Exchange decimal values using JSON

Use the decimal marshaler and unmarshal directly to the decimal type.
Do not decode values to `float64` first if all decimal digits must survive.
The inherited `Optimus` JSON contract maps `null` to zero, not to a missing
value.

**Run:** `go run ./examples/04_json` — [code and output](../examples/04_json/README.md).

## 7. Work with nullable SQL DECIMAL

```go
var column optimus.SQLDecimal
if err := column.Scan("123.4500"); err != nil {
    log.Fatal(err)
}
fmt.Println(column.Valid, column.Decimal.StringFixed()) // true 123.4500
```

`SQLDecimal` distinguishes NULL from zero and rejects floating-point input.
For the older `dec128.Dec128` SQL scanner, use
`dec128.ScanDestination(&value)` to update the destination.

**Run:** `go run ./examples/05_sql` — [SQL guide](../examples/05_sql/README.md).

## 8. Convert between decimal widths safely

Use `Dec128Exact`, `Dec256Exact` or `Dec512Exact` for narrowing, and handle
the error rather than silently rounding. If rounding is allowed, do it as a
separate explicit operation before conversion.

**Run:** `go run ./examples/09_conversions` — [code and output](../examples/09_conversions/README.md).

## 9. Use an unsigned 128-bit integer

`dec128/uint128.FromString` returns an integer and an `errors.Error` **code**,
not an `optimus.Optimus` or an ordinary `error` value. `errors.None` denotes
success, and checked arithmetic reports overflow.

**Run:** `go run ./examples/12_uint128` — [code and output](../examples/12_uint128/README.md).

## 10. Calculate concurrently

Share read-only source decimals, assign separate result slots, and choose
explicit scales per operation. Avoid changing process-wide default precision
from concurrent workers.

**Run:** `go run ./examples/10_concurrency` — [code and output](../examples/10_concurrency/README.md).

## More examples

[Dec128](../examples/06_dec128/README.md) ·
[Dec256](../examples/07_dec256/README.md) ·
[Dec512](../examples/08_dec512/README.md) ·
[Existing handler pipeline](../examples/11_decimal_pipeline/README.md) ·
[All runnable examples](../examples/README.md)

Short snippets in this guide assume the imports and surrounding `main`
function shown by their linked programs. See [Decimal behavior](DECIMAL_BEHAVIOR.md)
for limits, NaN, scale, and rounding rules.
