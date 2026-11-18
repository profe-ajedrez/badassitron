# Choosing a decimal type

## What all four types share

Each value has a sign, a nonnegative integer coefficient, and a decimal
scale. A coefficient of 12345 at scale 2 means 123.45. Scale is the
number of digits after the decimal point in the stored representation,
not the total count of significant digits.

These types implement decimal arithmetic directly rather than converting
operands to `float64`. They can represent values exactly only within
their coefficient and scale limits.

## Comparison

| Type | Maximum scale | Coefficient capacity | What to choose |
| --- | --- | --- | --- |
| [`dec128.Dec128`](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec128#Dec128) | 19 | 128 bits | Existing fixed-width integrations |
| [`dec256.Dec256`](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec256#Dec256) | 77 | 256 bits | Larger fixed-width values |
| [`dec512.Dec512`](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec512#Dec512) | 154 | 512 bits | Largest supported fixed-width values |
| [`optimus.Optimus`](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus#Optimus) | 154 | Automatic 64, 128, 256, or 512 bits | General arithmetic and adaptive storage |

A wider type does not decide a legally correct rounding policy. Nor does
it make division by 3 exactly representable as a finite decimal.

## Fixed-width values

The Dec128, Dec256, and Dec512 APIs provide `FromString`,
`NewFromString`, `FromInt`, arithmetic, comparisons, and explicit
rounding. Each has different capacity and a defined failure behavior.
Use a checked parser for external data.

Each package includes these executable examples:

- [`dec128/example_doc01_test.go`](../dec128/example_doc01_test.go)
- [`dec256/example_doc01_test.go`](../dec256/example_doc01_test.go)
- [`dec512/example_doc01_test.go`](../dec512/example_doc01_test.go)

They parse `12.3400` and compare half-away with half-even rounding.

## Adaptive Optimus

Optimus picks the smallest supported coefficient width that can hold an
operation's result. The implementation can promote from 64 to 128, 256,
or 512 bits. Its `Width()` tells you the current storage tier; it should
not be used as an economic classification.

`String()` normalizes output and `StringFixed()` prints the stored scale.
`Equal` and `Compare` compare numeric values across representations.
A plain Go `==` comparison is not available for Optimus; use `Equal`.

For the full arithmetic contract, see [Optimus details](OPTIMUS.md).

## Conversions

Optimus provides `FromDec128`, `FromDec256`, and `FromDec512` to widen
fixed-width decimals. `Dec128Exact`, `Dec256Exact`, and `Dec512Exact`
return an error when narrowing cannot preserve the value.
Do not replace an exact narrowing check with an unnoticed rounding step.

Run [`ExampleOptimus_Dec128Exact`](../optimus/example_doc01_test.go)
or the complete [conversion program](../examples/09_conversions/README.md).

## Unsigned integer packages

`dec128/uint128`, `dec256/uint256`, and `dec512/uint512` expose their
own fixed-width unsigned operations. They report overflow rather than
silently wrapping on checked arithmetic. They are building blocks for
decimal coefficients, not themselves signed decimal types.
See [integer foundation](UNSIGNED_INTEGERS.md).
