# Concurrent use of Optimus

Optimus arithmetic returns new values. Multiple goroutines may read the
same source decimal and calculate separate results without mutating the
source. Use **one result destination per goroutine**.

## Use explicit precision

```go
one := optimus.FromInt(1)
seven := optimus.FromInt(7)
result := one.DivRound(seven, 4, optimus.RoundingHalfEven)
if result.IsNaN() {
    log.Fatal(result.ErrorDetails())
}
fmt.Println(result.StringFixed()) // 0.1429
```

The snippet assumes the imports in the
[complete concurrency example](../examples/10_concurrency/main.go), which
runs three calculations in parallel with scales 2, 4, and 6.

`SetDefaultPrecision` uses atomic storage, so readers and setters do not race
on the setting itself. However, it is **process-wide**: different operations
may observe different default scales if other goroutines update it. Atomicity
is not a multi-operation consistency guarantee. Prefer per-operation scale
and mode when reproducibility matters.

## Do not share writable destinations

`Scan`, `UnmarshalJSON`, and `UnmarshalText` write into their destinations.
A concurrent read or write of the **same destination** requires
synchronization. `SQLDecimal` destinations follow the same rule.

Large Optimus coefficients can share internal read-only storage after a
value copy; ordinary arithmetic does not mutate it. The guarantee depends
on treating values and exported package-level numeric globals as immutable.
Do not concurrently change exported lookup tables, globals, or a value used
by another goroutine.

## Check results

Concurrent arithmetic has the same numerical limits as sequential
arithmetic. Inspect `IsNaN()` and `ErrorDetails()`, and handle `inexact`
separately when using a `WithStatus` method. A race-free program does not
necessarily have a stable business rounding policy.

[Runnable concurrency example](../examples/10_concurrency/README.md) ·
[Decimal behavior](../docs/DECIMAL_BEHAVIOR.md)
