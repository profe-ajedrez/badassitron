# Imported Badassitron facade versus direct Jokruger

Both calls compiled in the same default build. 7 samples per route; medians in ns/op. Positive delta means facade overhead.

| Operation | Facade | Direct upstream | Overhead | Allocations (facade/upstream) |
|---|---:|---:|---:|---:|
| `Add` | 0.902 | 0.889 | +1.4% | (0, 0) / (0, 0) |
| `AddOverflow` | 0.885 | 0.885 | +0.1% | (0, 0) / (0, 0) |
| `MulExact` | 1.326 | 1.321 | +0.4% | (0, 0) / (0, 0) |
| `MulHighExact` | 2.642 | 2.632 | +0.4% | (0, 0) / (0, 0) |
| `MulNarrowExact` | 2.627 | 2.678 | -1.9% | (0, 0) / (0, 0) |
| `MulOverflow` | 1.990 | 1.988 | +0.1% | (0, 0) / (0, 0) |
| `ParseLong` | 40.650 | 39.880 | +1.9% | (0, 0) / (0, 0) |
| `ParseShort` | 6.961 | 6.374 | +9.2% | (0, 0) / (0, 0) |
| `QuoRem` | 6.754 | 6.590 | +2.5% | (0, 0) / (0, 0) |
| `QuoRem64` | 2.617 | 2.610 | +0.3% | (0, 0) / (0, 0) |
| `QuoRemZero` | 1.778 | 1.559 | +14.0% | (0, 0) / (0, 0) |
| `StringToBufWide` | 47.080 | 46.790 | +0.6% | (0, 0) / (0, 0) |
| `WideOverflow` | 2.436 | 2.221 | +9.7% | (0, 0) / (0, 0) |
| `WideQuoRem` | 9.196 | 9.200 | -0.0% | (0, 0) / (0, 0) |

Results are descriptive microbenchmarks, not a pass/fail threshold or an equivalence proof.
Compare consumer workloads separately. Historical baselines remain in M3-F official evidence.
