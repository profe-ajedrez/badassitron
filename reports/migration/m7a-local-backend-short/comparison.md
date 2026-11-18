# M7-A — imported facade versus direct Jokruger

Both calls compiled in the same default build. 2 samples per route; medians in ns/op. Positive delta means facade overhead.

| Operation | Facade | Direct upstream | Overhead | Allocations (facade/upstream) |
|---|---:|---:|---:|---:|
| `Add` | 2.695 | 1.749 | +54.0% | (0, 0) / (0, 0) |
| `AddOverflow` | 1.877 | 1.445 | +29.9% | (0, 0) / (0, 0) |
| `MulExact` | 3.242 | 2.734 | +18.6% | (0, 0) / (0, 0) |
| `MulHighExact` | 4.531 | 5.065 | -10.5% | (0, 0) / (0, 0) |
| `MulNarrowExact` | 4.419 | 4.357 | +1.4% | (0, 0) / (0, 0) |
| `MulOverflow` | 3.679 | 3.380 | +8.9% | (0, 0) / (0, 0) |
| `ParseLong` | 73.960 | 86.265 | -14.3% | (0, 0) / (0, 0) |
| `ParseShort` | 14.820 | 13.475 | +10.0% | (0, 0) / (0, 0) |
| `QuoRem` | 16.905 | 17.845 | -5.3% | (0, 0) / (0, 0) |
| `QuoRem64` | 5.683 | 5.008 | +13.5% | (0, 0) / (0, 0) |
| `QuoRemZero` | 2.625 | 2.511 | +4.5% | (0, 0) / (0, 0) |
| `StringToBufWide` | 74.070 | 83.595 | -11.4% | (0, 0) / (0, 0) |
| `WideOverflow` | 2.848 | 4.045 | -29.6% | (0, 0) / (0, 0) |
| `WideQuoRem` | 18.425 | 14.795 | +24.5% | (0, 0) / (0, 0) |

Results are descriptive microbenchmarks, not a pass/fail threshold or an equivalence proof.
Compare consumer workloads separately. Historical baselines remain in M3-F official evidence.
