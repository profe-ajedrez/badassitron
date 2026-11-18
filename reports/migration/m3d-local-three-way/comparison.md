# Three-way uint128 benchmark comparison

Medians in ns/op; negative delta is faster. Same source tests, compiled under both build configurations.

| Operation | Historical facade | Direct Jokruger (historical build) | M3 adapter | Direct Jokruger (M3 build) | Adapter vs historical | Adapter vs direct Jokruger (M3 build) |
|---|---:|---:|---:|---:|---:|---:|
| `Add` | 1.146 | 1.156 | 1.528 | 1.176 | +33.3% | +29.9% |
| `MulExact` | 1.680 | 1.670 | 1.985 | 1.776 | +18.2% | +11.8% |
| `MulNarrowExact` | 2.563 | 2.835 | 3.222 | 2.963 | +25.7% | +8.7% |
| `MulHighExact` | 2.861 | 2.873 | 3.207 | 3.090 | +12.1% | +3.8% |
| `MulOverflow` | 2.289 | 2.363 | 2.687 | 2.321 | +17.4% | +15.8% |
| `QuoRem` | 10.920 | 10.670 | 10.300 | 10.220 | -5.7% | +0.8% |
| `WideQuoRem` | 11.450 | 11.920 | 11.960 | 11.680 | +4.5% | +2.4% |
| `AddOverflow` | 1.389 | 1.405 | 1.356 | 1.225 | -2.4% | +10.7% |
| `QuoRem64` | 4.025 | 3.720 | 3.539 | 3.500 | -12.1% | +1.1% |
| `QuoRemZero` | 1.887 | 1.968 | 3.258 | 2.337 | +72.7% | +39.4% |
| `WideOverflow` | 2.700 | 2.639 | 3.831 | 2.619 | +41.9% | +46.3% |

All eleven operations use variable inputs from a 32-element array and preserve output and status in package sinks.
Comparison is not proof of the underlying cause; inspect inlining and assembly, and repeat with official Go 1.27.1.
Errors/overflow paths are not represented equally across all operations; avoid extrapolating beyond the given operands.
