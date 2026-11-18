# Three-way uint128 benchmark comparison

Medians in ns/op; negative delta is faster. Same source tests, compiled under both build configurations.

| Operation | Historical facade | Direct Jokruger (historical build) | M3 adapter | Direct Jokruger (M3 build) | Adapter vs historical | Adapter vs direct Jokruger (M3 build) |
|---|---:|---:|---:|---:|---:|---:|
| `Add` | 1.221 | 1.243 | 1.541 | 1.218 | +26.2% | +26.5% |
| `MulExact` | 1.890 | 1.975 | 2.052 | 1.874 | +8.6% | +9.5% |
| `MulNarrowExact` | 2.787 | 3.271 | 3.639 | 3.171 | +30.6% | +14.8% |
| `MulHighExact` | 3.197 | 3.141 | 3.532 | 3.189 | +10.5% | +10.8% |
| `MulOverflow` | 2.523 | 2.487 | 3.035 | 2.533 | +20.3% | +19.8% |
| `QuoRem` | 10.890 | 10.640 | 11.350 | 10.600 | +4.2% | +7.1% |
| `WideQuoRem` | 12.670 | 12.760 | 12.920 | 12.620 | +2.0% | +2.4% |

All seven operations use variable inputs from a 32-element array and preserve output and status in package sinks.
Comparison is not proof of the underlying cause; inspect inlining and assembly, and repeat with official Go 1.27.1.
Errors/overflow paths are not represented equally across all operations; avoid extrapolating beyond the given operands.
