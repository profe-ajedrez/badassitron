# Imported Badassitron facade versus direct Jokruger

Both calls compiled in the same default build. 7 samples per route; medians in ns/op. Positive delta means facade overhead.

| Operation | Facade | Direct upstream | Overhead | Allocations (facade/upstream) |
|---|---:|---:|---:|---:|
| `Add` | 0.879 | 0.861 | +2.0% | (0, 0) / (0, 0) |
| `AddOverflow` | 0.879 | 0.893 | -1.5% | (0, 0) / (0, 0) |
| `MulExact` | 1.307 | 1.282 | +2.0% | (0, 0) / (0, 0) |
| `MulHighExact` | 2.624 | 2.602 | +0.8% | (0, 0) / (0, 0) |
| `MulNarrowExact` | 2.597 | 2.608 | -0.4% | (0, 0) / (0, 0) |
| `MulOverflow` | 1.950 | 1.963 | -0.7% | (0, 0) / (0, 0) |
| `ParseLong` | 39.280 | 37.910 | +3.6% | (0, 0) / (0, 0) |
| `ParseShort` | 6.884 | 6.170 | +11.6% | (0, 0) / (0, 0) |
| `QuoRem` | 6.635 | 6.456 | +2.8% | (0, 0) / (0, 0) |
| `QuoRem64` | 2.627 | 2.570 | +2.2% | (0, 0) / (0, 0) |
| `QuoRemZero` | 1.759 | 1.538 | +14.4% | (0, 0) / (0, 0) |
| `StringToBufWide` | 45.910 | 45.640 | +0.6% | (0, 0) / (0, 0) |
| `WideOverflow` | 2.402 | 2.167 | +10.8% | (0, 0) / (0, 0) |
| `WideQuoRem` | 9.117 | 9.294 | -1.9% | (0, 0) / (0, 0) |

Results are descriptive microbenchmarks, not a pass/fail threshold or an equivalence proof.
Compare consumer workloads separately. Historical baselines remain in M3-F official evidence.
