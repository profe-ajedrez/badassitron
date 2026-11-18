# Imported Badassitron facade versus direct Jokruger

Both calls compiled in the same default build. 7 samples per route; medians in ns/op. Positive delta means facade overhead.

| Operation | Facade | Direct upstream | Overhead | Allocations (facade/upstream) |
|---|---:|---:|---:|---:|
| `Add` | 1.089 | 0.877 | +24.2% | (0, 0) / (0, 0) |
| `AddOverflow` | 0.875 | 0.880 | -0.6% | (0, 0) / (0, 0) |
| `MulExact` | 1.307 | 1.297 | +0.8% | (0, 0) / (0, 0) |
| `MulHighExact` | 2.601 | 2.601 | +0.0% | (0, 0) / (0, 0) |
| `MulNarrowExact` | 2.604 | 2.612 | -0.3% | (0, 0) / (0, 0) |
| `MulOverflow` | 1.967 | 1.968 | -0.1% | (0, 0) / (0, 0) |
| `ParseLong` | 39.980 | 39.410 | +1.4% | (0, 0) / (0, 0) |
| `ParseShort` | 7.001 | 6.258 | +11.9% | (0, 0) / (0, 0) |
| `QuoRem` | 6.706 | 6.554 | +2.3% | (0, 0) / (0, 0) |
| `QuoRem64` | 2.571 | 2.560 | +0.4% | (0, 0) / (0, 0) |
| `QuoRemZero` | 1.760 | 1.549 | +13.6% | (0, 0) / (0, 0) |
| `StringToBufWide` | 46.590 | 45.980 | +1.3% | (0, 0) / (0, 0) |
| `WideOverflow` | 2.408 | 2.197 | +9.6% | (0, 0) / (0, 0) |
| `WideQuoRem` | 9.171 | 9.098 | +0.8% | (0, 0) / (0, 0) |

Results are descriptive microbenchmarks, not a pass/fail threshold or an equivalence proof.
Compare consumer workloads separately. Historical baselines remain in M3-F official evidence.
