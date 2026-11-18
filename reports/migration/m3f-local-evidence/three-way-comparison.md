# Three-way uint128 benchmark comparison

Medians in ns/op; negative delta is faster. Same source tests, compiled under both build configurations.

| Operation | Historical facade | Direct Jokruger (historical build) | M3 adapter | Direct Jokruger (M3 build) | Adapter vs historical | Adapter vs direct Jokruger (M3 build) |
|---|---:|---:|---:|---:|---:|---:|
| `Add` | 1.491 | 1.497 | 2.753 | 1.717 | +84.6% | +60.3% |
| `MulExact` | 2.574 | 2.545 | 2.936 | 2.510 | +14.1% | +17.0% |
| `MulNarrowExact` | 3.885 | 4.406 | 4.670 | 4.316 | +20.2% | +8.2% |
| `MulHighExact` | 3.888 | 4.281 | 4.333 | 4.319 | +11.4% | +0.3% |
| `MulOverflow` | 3.381 | 3.518 | 3.685 | 3.393 | +9.0% | +8.6% |
| `QuoRem` | 16.690 | 17.030 | 17.320 | 18.100 | +3.8% | -4.3% |
| `WideQuoRem` | 14.600 | 15.180 | 20.710 | 16.840 | +41.8% | +23.0% |
| `AddOverflow` | 1.479 | 1.496 | 1.927 | 1.671 | +30.3% | +15.3% |
| `QuoRem64` | 3.922 | 3.954 | 4.283 | 4.321 | +9.2% | -0.9% |
| `QuoRemZero` | 2.432 | 2.429 | 2.630 | 2.575 | +8.1% | +2.1% |
| `WideOverflow` | 3.997 | 4.494 | 2.913 | 4.165 | -27.1% | -30.1% |

All eleven operations use variable inputs from a 32-element array and preserve output and status in package sinks.
Comparison is not proof of the underlying cause; inspect inlining and assembly, and repeat with official Go 1.27.1.
Errors/overflow paths are not represented equally across all operations; avoid extrapolating beyond the given operands.
