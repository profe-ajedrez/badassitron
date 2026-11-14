# Optimus user guide

This guide describes the Optimus v8 source snapshot. Earlier stage and audit
files document implementation history. For a first example, see the
[README](../README.md).

## What Optimus stores

An `Optimus` value has a coefficient, a sign, and a decimal scale. The
coefficient is an unsigned integer. Its internal storage grows from 64 to
128, 256, or 512 bits as needed. Small coefficients use inline storage.

A value with coefficient 12345 and scale 2 represents 123.45. Its stored
scale is returned by `Precision()` and `Exponent()`. Both names refer to
fractional decimal places, not significant digits. The largest scale is 154.
`Width()` reports the current coefficient tier. Do not choose business logic
based on the tier because it is a storage detail.

`Canonical()` removes trailing decimal zeros while preserving the numeric
value. `Equal()` and `Compare()` compare numeric values even when stored
scales or storage tiers differ. `String()` removes trailing fractional zeros.
`StringFixed()` keeps the current stored scale.

## Construct values

```go
x := optimus.FromString("123.4500")
y := optimus.FromInt(7)
z, err := optimus.NewFromString("0.125")
```

Use `NewFromString` for input validation. `FromString` returns `NaN` for
invalid text. The inherited empty string behavior maps an empty input to
zero. `FromFloat64` first formats a binary floating point input as decimal
text. Do not use it to recover decimal digits that were already lost when
converting an exact value to floating point.

The constructor `New` takes a 512 bit coefficient, an explicit scale, and a
sign. Smaller coefficient constructors are available. Explicit constructors
reject a scale above 154 instead of rounding it to fit.

## Arithmetic policy

`Add`, `Sub`, and `Mul` retain an exact result whenever its coefficient and
scale can be represented. Otherwise Optimus removes exact trailing zeros
when useful and rounds half even once at the finest scale that fits. An
integer magnitude larger than the maximum coefficient produces `NaN` with
an overflow reason. Ordinary arithmetic does not return an inexact flag.

`Div` keeps terminating decimal quotients exactly when representable.
Repeating decimal quotients use half even rounding and a working precision
that starts at 19 fractional places. `Sqrt` preserves exact square roots
and rounds irrational results by the same default policy. `PowInt` supports
integer exponents, including interval checked extreme powers. Some near
midpoint extreme powers return a precision error rather than an unproved
rounded result.

`QuoRem` returns an integer quotient and a remainder with the dividend sign.
The quotient truncates toward zero. `Mod` returns the corresponding remainder
and may succeed even when an integer quotient would be too large to return.
Division by zero produces `NaN`.

## Choose a rounding operation

| Operation | Behavior |
| --- | --- |
| `Round(p)` | Nearest value, ties away from zero |
| `RoundBank(p)` | Nearest value, ties to even |
| `RoundDown(p)` | Toward negative infinity |
| `RoundUp(p)` | Toward positive infinity |
| `Trunc(p)` | Discard fractional digits toward zero |
| `Rescale(p)` | Set scale, truncating when digits must be discarded |
| `RescaleRound(p, mode)` | Set scale using the requested rounding mode |
| `MulRound`, `DivRound`, `SqrtRound` | Calculate and quantize at an explicit scale |

Use the `WithStatus` variants when discarded digits matter. A returned
`true` means nonzero information was discarded. The boolean does not mean
that the result is invalid. Check `IsNaN()` and `ErrorDetails()` for errors.
An exact but unrepresentable result can be invalid even when the inexact
boolean is false.

```go
one := optimus.FromInt(1)
three := optimus.FromInt(3)
value, inexact := one.DivRoundWithStatus(
    three, 2, optimus.RoundingHalfEven,
)
```

Explicit rounding methods use their requested scale and mode. They do not
silently reduce the scale to make a result fit. `Round` and `Trunc` with a
requested scale at or above the current scale leave the value unchanged.
`RescaleRound` can increase the stored scale and may overflow the coefficient.

## Precision and concurrency

The initial default working precision is 19. `SetDefaultPrecision` accepts
scales from 0 to 154 and panics for values above the maximum. Its setting is
process wide and stored atomically. Another goroutine can change it between
operations without a data race. A group of arithmetic operations is not a
precision transaction. Choose explicit rounding APIs for reproducible
concurrent calculations.

Optimus arithmetic treats inputs as immutable values. A copied wide value
may share internal read only coefficient storage. Never modify exported
numeric globals or lookup slices concurrently with code that reads them.
See [concurrency](../optimus/CONCURRENCY.md).

## NaN, zero, text, and JSON

The zero value of `Optimus` represents numeric zero. Negative zero has no
persistent negative sign. `IsZero()` is false for NaN. `NaN` values compare
equal to each other regardless of reason and sort before ordinary numbers.
This ordering is an inherited library contract, not an IEEE numeric order.

`String()` emits `NaN` for invalid values. JSON writes decimals as quoted
strings, including the quoted text `NaN`. Text and JSON parsers reject `NaN`
as ordinary decimal input. Therefore a NaN value cannot round trip through
text or JSON. JSON `null` maps to numeric zero. These compatibility rules are
preserved in v8.

## Database SQL

`Optimus` implements the older database interfaces. Its `Scan` accepts text,
bytes, integers, and floating point values. SQL NULL becomes numeric zero.
`Value` returns decimal text, including `NaN` for an invalid value. This
path is kept for compatibility, not for strict loss free database input.

Use `SQLDecimal` for strict database values. It implements `sql.Scanner` and
`driver.Valuer`. A NULL scan resets `Valid` to false. A valid decimal is
stored in `Decimal` with `Valid` true. It accepts exact decimal text and
integers. It rejects binary floating point input and decimal strings that
cannot be represented exactly. A failed scan leaves the destination as it
was. A valid write returns decimal text. An invalid decimal write returns
an error. A NULL write returns nil.

```go
var scanned optimus.SQLDecimal
if err := scanned.Scan("123.4500"); err != nil {
    return err
}
if scanned.Valid {
    fmt.Println(scanned.Decimal.StringFixed())
}
```

A database driver must return exact decimal text or an integer for the strict
scan contract. Database column precision and formatting also matter. The
included SQL integration tests use a local in memory driver. They do not
establish compatibility with a particular external database driver.

## Convert to other decimal types

`FromDec128`, `FromDec256`, and `FromDec512` import existing decimal values.
`Dec128Exact`, `Dec256Exact`, and `Dec512Exact` attempt exact output
conversions. If necessary, a conversion removes trailing zeros first.
A value that cannot fit the requested target returns an error rather than
being rounded. The decimal types remain distinct Go types.

The unsigned encoders such as `EncodeToUint128` rescale first and can
truncate fractional digits. They are not substitutes for `Dec128Exact`.
Use `Int` and `Int64` only when an integer conversion is intended.
`InexactFloat64` makes an explicit binary floating point conversion.

## Choosing the other packages

`Dec128`, `Dec256`, and `Dec512` keep a fixed coefficient width. Their
ordinary division and square root policies are not the same as the adaptive
Optimus policy. In particular, fixed width division and square root use
truncation toward zero in their ordinary methods. Use explicit rounding
methods whenever a shared rounding rule is required across packages.

## Known limits in v8

- A single 512 bit coefficient and scale 154 cannot represent every rational
  value. Periodic division and irrational roots require rounding.
- The Optimus v8 correction detects overflowing 128 bit scale alignment in
  quotient and remainder calculations. It uses a slower exact fallback.
- Legacy `uint128.Mul` has a separately recorded overflow reporting issue.
  Legacy `uint128.FromBigInt` can mutate the input argument. These are not
  changed by the Optimus v8 fix.
- Older NaN text round trips and SQL scanning rules remain unchanged.
- The original audit identified unrelated compile errors in
  `withfloat64/internal/math.go`; these were subsequently corrected.
- The v8 audit ran with an external Go 1.23 compatibility modfile, not with
  the declared Go 1.24 toolchain.

See [the v8 audit](../OPTIMUS_V8_PHASE_C_PRECISION_AUDIT.md) for the
reproduction, regression tests, benchmark results, and validation scope.

## Verify and navigate source

```sh
go test ./optimus/... ./dec128/... ./dec256/... ./dec512/...
go test -race ./optimus/...
go vet ./optimus/...
go doc -all ./optimus
```

GoDoc describes the public API directly from the code. The repository audit
reports preserve the results and limits of earlier campaigns.
