# Calculate with Dec512

Run this from the repository root:

```sh
go run ./examples/08_dec512
```

Read [main.go](main.go) for the complete program.

Dec512 stores a fixed 512 bit coefficient and up to 154 fractional places.
The first value is a one followed by 80 zeros. It is too large for Dec256
and still small enough for Dec512. Adding seven preserves every digit.

DecodeFromUint64(1, 100) builds a positive value with coefficient 1 and
scale 100. That means a decimal point, 99 zeros, and a final 1. A small
coefficient can represent a very small number when the scale is large.

SqrtRoundWithStatus calculates the square root of two to 30 fractional
places. The exact root has more digits, so the status is true. The result
is still a valid rounded decimal.

The tiny value cannot be converted exactly to Dec256 because its nonzero
digit is beyond Dec256's maximum scale of 77. Dec256Exact returns an error
instead of silently turning the value into zero.

Expected output:

```text
Large total: 100000000000000000000000000000000000000000000000000000000000000000000000000000007
Tiny scale: 100
Tiny coefficient: 1
Square root of two: 1.414213562373095048801688724210
Square root lost digits: true
Exact Dec256 conversion rejected: true
```

Back to the [example index](../README.md) or the
[ten-minute guide](../../docs/GETTING_STARTED.md).
