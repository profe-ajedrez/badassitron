# Three-way uint128 benchmark comparison

Medians in ns/op; negative delta is faster. Same source tests, compiled under both build configurations.

| Operation | Historical facade | Direct Jokruger (historical build) | M3 adapter | Direct Jokruger (M3 build) | Adapter vs historical | Adapter vs direct Jokruger (M3 build) |
|---|---:|---:|---:|---:|---:|---:|
| `Add` | 0.679 | 0.900 | 1.125 | 0.909 | +65.6% | +23.7% |
| `MulExact` | 1.339 | 1.343 | 1.565 | 1.323 | +16.9% | +18.3% |
| `MulNarrowExact` | 2.461 | 2.645 | 3.128 | 2.649 | +27.1% | +18.1% |
| `MulHighExact` | 2.458 | 2.646 | 3.127 | 2.642 | +27.2% | +18.4% |
| `MulOverflow` | 1.992 | 2.028 | 2.215 | 1.994 | +11.2% | +11.1% |
| `QuoRem` | 6.543 | 6.838 | 7.686 | 6.923 | +17.5% | +11.0% |
| `WideQuoRem` | 9.761 | 9.234 | 10.290 | 9.286 | +5.4% | +10.8% |
| `AddOverflow` | 0.786 | 0.910 | 1.115 | 0.894 | +41.8% | +24.7% |
| `QuoRem64` | 2.621 | 2.627 | 2.647 | 2.627 | +1.0% | +0.8% |
| `QuoRemZero` | 1.800 | 1.574 | 2.905 | 1.576 | +61.4% | +84.3% |
| `WideOverflow` | 2.215 | 2.653 | 3.761 | 2.509 | +69.8% | +49.9% |

All eleven operations use variable inputs from a 32-element array and preserve output and status in package sinks.
Comparison is not proof of the underlying cause; inspect inlining and assembly, and repeat with official Go 1.27.1.
Errors/overflow paths are not represented equally across all operations; avoid extrapolating beyond the given operands.
