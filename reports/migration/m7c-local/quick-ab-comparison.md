# Release adapter iteration comparison

3 samples per route. Each build is normalized to its own direct Jokruger results.
Negative overhead delta (percentage points) indicates improvement. This is descriptive, not a statistical guarantee.

| Operation | M7-A facade | M7-A direct | M7-A overhead | M7-C facade | M7-C direct | M7-C overhead | Delta (pp) |
|---|---:|---:|---:|---:|---:|---:|---:|
| `Add` | 2.583 | 1.640 | +57.5% | 1.754 | 1.790 | -2.0% | -59.5 |
| `MulExact` | 2.759 | 2.520 | +9.5% | 2.741 | 2.721 | +0.7% | -8.7 |
| `MulOverflow` | 3.604 | 3.437 | +4.9% | 3.479 | 3.589 | -3.1% | -7.9 |
| `QuoRem` | 17.770 | 16.510 | +7.6% | 13.700 | 17.840 | -23.2% | -30.8 |
| `WideQuoRem` | 19.000 | 16.980 | +11.9% | 15.540 | 16.490 | -5.8% | -17.7 |

The builds may use different code generation and run at different CPU frequencies. 
Use interleaved repetitions and consumer benchmarks before accepting a performance claim.
