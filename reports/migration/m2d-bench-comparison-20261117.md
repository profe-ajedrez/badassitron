# M6 public API benchmark comparison

Comparison is descriptive, not a confidence interval or release approval.
Both modes were benchmarked on the same host and toolchain, using the same inputs.

Environment: goversion=go version go1.23.2 linux/amd64, goos=linux, goarch=amd64, cpu=AMD EPYC 9V74 80-Core Processor

Provisional review trigger: greater than +10% (M2 slower).
Required samples per operation and variant: 5.

| Public operation | Default median ns/op | M2 median ns/op | Delta | Samples | B/op (default → M2) | Review |
|---|---:|---:|---:|---:|---:|---|
| `Add/MixedScale` | 11.59 | 10.29 | -11.2% | 5/5 | 0 → 0 | within sample threshold |
| `Add/NearOverflow` | 19.48 | 10.98 | -43.6% | 5/5 | 0 → 0 | within sample threshold |
| `Add/OppositeSign` | 3.84 | 3.92 | +2.0% | 5/5 | 0 → 0 | within sample threshold |
| `Add/Small` | 4.01 | 3.60 | -10.2% | 5/5 | 0 → 0 | within sample threshold |
| `Compare/MixedScale` | 7.37 | 7.73 | +4.8% | 5/5 | 0 → 0 | within sample threshold |
| `Div/Exact` | 10.03 | 10.71 | +6.8% | 5/5 | 0 → 0 | within sample threshold |
| `Div/Repeating` | 10.34 | 10.44 | +1.0% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/Large` | 23.34 | 15.59 | -33.2% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/ScaleOverflow` | 18.76 | 12.67 | -32.5% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/Small` | 4.24 | 3.99 | -6.0% | 5/5 | 0 → 0 | within sample threshold |
| `Parse/Long` | 84.41 | 48.13 | -43.0% | 5/5 | 0 → 0 | within sample threshold |
| `Parse/Short` | 10.56 | 8.55 | -19.0% | 5/5 | 0 → 0 | within sample threshold |
| `Rescale/RoundHalfEven` | 10.53 | 5.44 | -48.3% | 5/5 | 0 → 0 | within sample threshold |
| `Rescale/Truncate` | 4.13 | 3.93 | -4.9% | 5/5 | 0 → 0 | within sample threshold |
| `Rounding/DivRound` | 14.62 | 14.74 | +0.8% | 5/5 | 0 → 0 | within sample threshold |
| `Rounding/MulRound` | 15.10 | 12.65 | -16.2% | 5/5 | 0 → 0 | within sample threshold |
| `SQL/ScanInto` | 23.62 | 21.76 | -7.9% | 5/5 | 0 → 0 | within sample threshold |
| `Sqrt/FullWidth` | 175.40 | 144.20 | -17.8% | 5/5 | 0 → 0 | within sample threshold |
| `String/Canonical` | 42.54 | 45.37 | +6.7% | 5/5 | 16 → 16 | within sample threshold |
| `String/Fixed` | 43.02 | 43.01 | -0.0% | 5/5 | 16 → 16 | within sample threshold |

Review cases (0): none

This does **not** prove equivalence for arbitrary operands or engines.
Use Go 1.27.1 and the official v1.4.0 module before release approval.
