# Optimus concurrency contract

This document records the Phase C concurrency contract from Optimus v7,
which remains applicable to v8. It adds no
public API and does not alter the existing numerical contracts.

## Shared values

An `Optimus` value may be copied and used concurrently for read-only operations
and arithmetic. Arithmetic methods return new values; they do not modify their
receivers or other operands. Coefficients at the 256- and 512-bit tiers are
stored indirectly, so copying an `Optimus` can share the same underlying
`uint512.Uint512`. That storage is private and must remain immutable after
construction. `Coefficient()` returns a value, while `coefBig()` and `core()`
construct independent `math/big.Int` objects.

The value is **immutable by convention**, not mechanically frozen. Code within
the `optimus` package must not mutate the private `wide` pointer's target after
construction. Existing exported package variables (such as `Zero` and `One`)
and exported byte slices are writable by consumers. Concurrent application
writes to those variables are **not** protected by this contract. The same
caution applies to mutable exported lookup tables in the underlying integer
packages. Treat these globals as read-only after initialization.

## Precision configuration

`SetDefaultPrecision` stores the requested value using `atomic.Uint32`. There
is no data race between the setter and operations that read the default.
However, precision is process-wide and dynamic. If one goroutine changes it
while another evaluates a non-terminating operation, the result may reflect
any applicable setting observed during that call. A sequence of calculations
is not a precision transaction and need not use the same setting throughout.

When an operation provides an explicit precision and rounding mode (such as
`DivRound`, `SqrtRound`, `RescaleRound`, or their `WithStatus` variants), use
those methods if the application requires precision independent of concurrent
configuration changes. Other operations (including some `PowInt` and `Sqrt`
paths) consult the global setting by design. Coordinate changes to that
setting at the application level when deterministic multi-operation results
are required.

## SQL and serialization

`(Optimus).Value`, `SQLDecimal.Value`, `String`, `MarshalJSON` and `MarshalText`
are read-only. They can run concurrently on the same immutable source. Each
`Scan`, `UnmarshalJSON` or `UnmarshalText` call writes to its destination:
concurrent writes to the **same** `Optimus` or `SQLDecimal` destination (or a
read concurrent with a write to that destination) require external
synchronization. Scanning into distinct destinations is safe.

These concurrency guarantees do not imply that a third-party SQL driver
preserves decimal precision. The exact SQL conversion rules are covered
separately in the Phase C SQL audit.

## Verification performed

`phase_c_concurrency_test.go` covers shared 64-, 256- and 512-bit operands,
indirect coefficient aliases, independent `math/big` conversions, bounded
`PowInt` with extreme signed exponents, mixed arithmetic, rounding, SQL,
serialization, and concurrent `SetDefaultPrecision` updates. The tests compare
complete decimal state (coefficient, scale, sign, error and tier) with
precomputed sequential outcomes. Parallel benchmarks provide reference
measurements for small and wide addition.

The race detector is a necessary part of this audit, but no finite test suite
proves every possible interleaving or every third-party integration safe.
