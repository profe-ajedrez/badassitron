# Three-way uint128 benchmark comparison

Medians in ns/op; negative delta is faster. Same source tests, compiled under both build configurations.

| Operation | Historical facade | Direct Jokruger (historical build) | M3 adapter | Direct Jokruger (M3 build) | Adapter vs historical | Adapter vs direct Jokruger (M3 build) |
|---|---:|---:|---:|---:|---:|---:|
| `Add` | 1.315 | 1.218 | 1.585 | 1.298 | +20.5% | +22.1% |
| `MulExact` | 1.833 | 1.761 | 2.126 | 2.073 | +16.0% | +2.6% |
| `MulOverflow` | 2.521 | 2.468 | 3.854 | 2.542 | +52.9% | +51.6% |
| `QuoRem` | 10.930 | 9.993 | 11.490 | 10.940 | +5.1% | +5.0% |
| `WideQuoRem` | 11.490 | 12.330 | 13.410 | 12.870 | +16.7% | +4.2% |

All five operations use variable inputs from a 32-element array and preserve output and status in package sinks.
Comparison is not proof of the underlying cause; inspect inlining and assembly, and repeat with official Go 1.27.1.
Errors/overflow paths are not represented equally across all operations; avoid extrapolating beyond the given operands.
