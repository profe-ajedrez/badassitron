# Precision, rounding, and errors

## Exact decimal values are still bounded

Text input such as `0.1` can be stored exactly without converting it to
binary floating point. This does not make every arithmetic result exact.
An infinite decimal expansion such as 1 divided by 3 must be rounded if
it is returned as a finite decimal.

Coefficient width and fractional scale are independent limits. A result
may not fit even when its scale is small. Some ordinary arithmetic paths
round an otherwise unrepresentable result, and some return a NaN/failure
state. Review the chosen type's implementation contract and check results.

## The basic rounding modes

| API or mode | What happens on an exact tie |
| --- | --- |
| `Round` or `RoundingHalfAwayFromZero` | Move away from zero |
| `RoundBank` or `RoundingHalfEven` | Keep the nearest even final digit |
| `RoundingHalfTowardZero` | Move toward zero |
| `RoundingTowardZero` | Discard additional fractional digits |
| `RoundingAwayFromZero` | Move outward |
| `RoundingFloor` | Move toward negative infinity |
| `RoundingCeil` | Move toward positive infinity |

For instance, 2.5 rounded to an integer is 3 with half-away and 2 with
half-even. With the same modes, -2.5 becomes -3 or -2.

See [`ExampleOptimus_Round`](../optimus/example_doc01_test.go).

## Explicit precision

`DivRound`, `MulRound`, `SqrtRound`, and `RescaleRound` let the caller
choose the output scale and rounding mode. Their `WithStatus` variants
return `(value, inexact)` where the boolean reports discarded nonzero
information. It does not indicate that the result is NaN and is not
an error object.

```go
value := optimus.FromString("1.235")
rounded, inexact := value.RescaleRoundWithStatus(
    2, optimus.RoundingHalfEven,
)
fmt.Println(rounded, inexact) // 1.24 true
```

This fragment is executed by
[`ExampleOptimus_RescaleRoundWithStatus`](../optimus/example_doc01_test.go).

`Rescale` and `Trunc` use truncation toward zero. Explicitly select the
mode when the calculation requires another policy.

## Default precision

`SetDefaultPrecision` changes a process-wide setting used by selected
non-terminating or irrational operations, including default division and
square root paths. Setter changes are made atomically but are not a
multi-operation precision transaction. Separate concurrent calculations
can observe different settings during a global update.

Prefer per-operation scale and mode for reproducible concurrent work.
Do not change default precision separately for each request or goroutine.
See [Optimus concurrency](../optimus/CONCURRENCY.md).

## Parsing and arithmetic failure

`NewFromString` returns a value and an error. This is generally the
right choice for external input.
`FromString` returns a value, and malformed input may produce NaN.
Operations may also produce NaN due to overflow, division by zero, or
precision constraints. Check `IsNaN()` and inspect `ErrorDetails()`.

```go
value := optimus.FromString("invalid")
if value.IsNaN() {
    fmt.Println(value.ErrorDetails())
}
```

Do not substitute zero for a failed calculation. Do not use an inexact
flag as a replacement for numeric error checks.

## Deterministic financial results

A financial rule should select its arithmetic scale and rounding mode
explicitly, document the point where rounding occurs, and test both
normal and boundary values. A tax framework must define these policies;
Badassitron cannot infer them from the numbers alone.

See the [rounding example program](../examples/02_rounding/README.md)
and [invoice example](../examples/03_invoice/README.md).
