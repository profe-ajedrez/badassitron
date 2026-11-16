# Decimal behavior: the decisions that affect results

This page is a quick reference for **current code** in the decimal packages.
It describes mechanics, not accounting or tax regulations. For working code,
read [Getting started](GETTING_STARTED.md).

## Scale is not significant-digit precision

A decimal is stored as `coefficient × 10^(-scale)` with a sign. `12345` at
scale 2 is `123.45`. The methods `Precision()` and, where available,
`Exponent()` refer to *scale*. They do not count significant digits.

| Type | Coefficient size | Maximum scale | Overflow behavior for ordinary exact arithmetic |
| --- | --- | --- | --- |
| `dec128.Dec128` | 128 bits | 19 | Returns NaN if the result cannot be represented exactly. |
| `dec256.Dec256` | 256 bits | 77 | Returns NaN if the result cannot be represented exactly. |
| `dec512.Dec512` | 512 bits | 154 | Returns NaN if the result cannot be represented exactly. |
| `optimus.Optimus` | Adaptive, max 512 bits | 154 | May round half-even to the finest representable scale; returns NaN when the integer magnitude cannot fit. |

Exact trailing zeros can sometimes be removed without changing the number.
Wider coefficients can hold larger integer magnitudes **or** more fractional
digits. Maximum scale by itself is not a guarantee that any magnitude fits.

## Ordinary operations and explicit rounding

`Add`, `Sub`, `Mul`, `Div`, and `Sqrt` do **not** all share one automatic
rounding policy. In particular:

| Operation | `optimus` | Fixed-width decimals |
| --- | --- | --- |
| `Add`, `Sub`, `Mul` | Exact when possible; may round half-even if a finite representable result requires fewer decimal places. | Exact when representable; otherwise NaN. |
| `Div` | Keeps representable terminating results exact; repeating quotients use a process-wide default working scale, initially 19. | Division uses truncation, with process-wide default precision (initially 19) and input-scale rules. |
| `Sqrt` | Preserves exact roots; otherwise uses its default rounding policy. | Ordinary square root uses truncation. |
| `*Round(…, scale, mode)` | Explicit rounding mode and target scale. | Explicit rounding mode and target scale. |

The explicit `MulRound`, `DivRound`, and `SqrtRound` families take a target
scale and rounding mode. The `WithStatus` variants also report whether any
nonzero digits were discarded. **A true status does not mean failure, and a
false status does not guarantee success**; check `IsNaN()`.

Available modes include toward zero, away from zero, floor, ceiling, nearest
with ties toward zero, nearest with ties away, and nearest with ties to even.
`Round` means ties **away** from zero; `RoundBank` means ties **to even**.

### Two different ways to request two places

```go
x := optimus.FromString("12.5")
fmt.Println(x.RoundBank(2).StringFixed()) // 12.5
fmt.Println(x.RescaleRound(2, optimus.RoundingHalfEven).StringFixed()) // 12.50
```

`RoundBank` does not increase the stored scale when no rounding is needed;
`RescaleRound` does. `String()` displays compact decimal text, whereas
`StringFixed()` retains the stored scale.

## Input and failure handling

| Situation | Recommended choice |
| --- | --- |
| Trusted literal / integer | `FromString("19.95")` or `FromInt(20)`. |
| User text with required validation | Check blank input; use `NewFromString` and the returned Go error. |
| Arithmetic error | Test `result.IsNaN()` and inspect `result.ErrorDetails()`. |
| Decimal precision from binary float | Avoid `FromFloat64` if the original exact decimal digits matter. |
| Narrowing a wide result | Use `Dec128Exact` / `Dec256Exact` / `Dec512Exact` and handle the error. |
| Nullable exact SQL DECIMAL | Use `optimus.SQLDecimal` and a driver that supplies decimal text or integers. |
| `Dec128` with `database/sql` | Use `dec128.ScanDestination(&d)` or `dec128.ScanInto(&d, src)`. |

A `NaN` decimal is an invalid value, not a Go `error`. The library maintains
historical NaN comparison and text behaviors; do not use those behaviors in
place of explicit error handling.

## Concurrency and shared settings

Decimal arithmetic returns new values and treats its operands as immutable.
Do not mutate exported lookup tables or shared result destinations without
coordination. `SetDefaultPrecision` is a process-wide atomic setting: its
atomicity avoids a data race, **but does not guarantee a consistent precision
across multiple operations if goroutines change it**. Prefer explicit-scale
operations in concurrent or reproducible calculations.

## What is not included

The numeric packages do not define a currency type, tax jurisdiction, or
mandatory rounding stages. Your application must decide when to round and
whether to reject inexact results. The older [calculation pipeline](CALCULATION_PIPELINE.md)
is separate from the arithmetic APIs and has incomplete paths.

See [examples](../examples/README.md), the [Optimus guide](OPTIMUS_STAGE.md),
and [Dec128 precision](../dec128/PRECISION.md) for details.
