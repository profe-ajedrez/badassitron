# Documentation map

Choose the shortest path that answers your question.

| If you need… | Read |
| --- | --- |
| To get useful results in ten minutes | [Getting started](GETTING_STARTED.md) |
| A quick guide to rounding, capacity, NaN and scales | [Decimal behavior](DECIMAL_BEHAVIOR.md) |
| Runnable code and expected output | [Example index](../examples/README.md) |
| Full adaptive decimal behavior | [Optimus guide](OPTIMUS_STAGE.md) |
| Fixed-width 128-bit scale and errors | [Dec128 precision](../dec128/PRECISION.md) |
| `database/sql` integration | [SQL example](../examples/05_sql/README.md) and [Dec128 scanning](../dec128/SQL_SCANNING.md) |
| A handler-based calculation pipeline | [Calculation pipeline](CALCULATION_PIPELINE.md) |
| Thread-safety limitations | [Concurrency](../optimus/CONCURRENCY.md) |
| Fixed-width designs | [Wide decimal stage](DECIMAL_WIDE_STAGE.md) |
| Integer implementation details | [Wide integer foundation](WIDE_INTEGER_FOUNDATION.md) |
| The V4 optimization, limits, and benchmarks | [FASTPATH V4](FASTPATH_V4.md) |

The `*_STAGE.md` documents include implementation history. Read the current
package GoDoc and runnable examples for a precise description of the checkout
in use. The project contains several historical reports; not every external
audit artifact is bundled in this source ZIP.

For API symbols use `go doc -all ./optimus`, `go doc -all ./dec128`, or
[pkg.go.dev](https://pkg.go.dev/github.com/profe-ajedrez/badassitron).
