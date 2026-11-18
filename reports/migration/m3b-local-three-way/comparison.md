# Three-way uint128 benchmark comparison

Medians in ns/op; negative delta is faster. Same source tests, compiled under both build configurations.

| Operation | Historical facade | Direct Jokruger (historical build) | M3 adapter | Direct Jokruger (M3 build) | Adapter vs historical | Adapter vs direct Jokruger (M3 build) |
|---|---:|---:|---:|---:|---:|---:|
| `Add` | 1.315 | 1.218 | 1.823 | 1.239 | +38.6% | +47.1% |
| `MulExact` | 1.833 | 1.761 | 2.233 | 1.834 | +21.8% | +21.8% |
| `MulOverflow` | 2.521 | 2.468 | 3.643 | 2.493 | +44.5% | +46.1% |
| `QuoRem` | 10.930 | 9.993 | 11.470 | 10.520 | +4.9% | +9.0% |
| `WideQuoRem` | 11.490 | 12.330 | 12.500 | 12.000 | +8.8% | +4.2% |

All five operations use variable inputs from a 32-element array and preserve output and status in package sinks.
Comparison is not proof of the underlying cause; inspect inlining and assembly, and repeat with official Go 1.27.1.
Errors/overflow paths are not represented equally across all operations; avoid extrapolating beyond the given operands.
