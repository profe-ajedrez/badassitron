# Round each line or round the total?

Run from the repository root:

```sh
go run ./examples/13_rounding_boundaries
```

[View the complete program](main.go).

The program sums `1.005`, `2.005`, and `3.005` using decimal arithmetic.
With **half-even rounding to two places**, rounding each input first yields a
different result from adding the exact inputs and rounding once:

Expected output:

```text
Line: 1.005 -> 1.00 inexact: true
Line: 2.005 -> 2.00 inexact: true
Line: 3.005 -> 3.00 inexact: true
Round each line: 6.00
Round the total: 6.02
```

Neither result is universally correct for invoices or taxes. The calling
application must choose its rounding stage and mode. `inexact: true` means
nonzero digits were discarded; it is not an error status. Always check
`IsNaN()` too.

[All examples](../README.md) · [Rounding and errors](../../docs/ROUNDING_AND_ERRORS.md)
