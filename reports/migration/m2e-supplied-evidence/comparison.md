# M6 public API benchmark comparison

Comparison is descriptive, not a confidence interval or release approval.
Both modes were benchmarked on the same host and toolchain, using the same inputs.

Environment: goversion=go version go1.27.1 linux/amd64, goos=linux, goarch=amd64, cpu=AMD Ryzen 7 7735HS with Radeon Graphics

Provisional review trigger: greater than +10% (M2 slower).
Required samples per operation and variant: 5.

| Public operation | Default median ns/op | M2 median ns/op | Delta | Samples | B/op (default → M2) | Review |
|---|---:|---:|---:|---:|---:|---|
| `Add/MixedScale` | 9.17 | 8.83 | -3.8% | 5/5 | 0 → 0 | within sample threshold |
| `Add/NearOverflow` | 13.97 | 9.37 | -33.0% | 5/5 | 0 → 0 | within sample threshold |
| `Add/OppositeSign` | 3.14 | 3.09 | -1.7% | 5/5 | 0 → 0 | within sample threshold |
| `Add/Small` | 3.17 | 3.11 | -2.0% | 5/5 | 0 → 0 | within sample threshold |
| `Compare/MixedScale` | 6.84 | 6.45 | -5.8% | 5/5 | 0 → 0 | within sample threshold |
| `Div/Exact` | 8.63 | 8.33 | -3.5% | 5/5 | 0 → 0 | within sample threshold |
| `Div/Repeating` | 8.75 | 8.33 | -4.8% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/Large` | 18.97 | 12.51 | -34.1% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/ScaleOverflow` | 16.37 | 10.25 | -37.4% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/Small` | 3.54 | 3.32 | -6.4% | 5/5 | 0 → 0 | within sample threshold |
| `Parse/Long` | 60.89 | 46.25 | -24.0% | 5/5 | 0 → 0 | within sample threshold |
| `Parse/Short` | 8.02 | 7.54 | -5.9% | 5/5 | 0 → 0 | within sample threshold |
| `Rescale/RoundHalfEven` | 9.74 | 4.89 | -49.7% | 5/5 | 0 → 0 | within sample threshold |
| `Rescale/Truncate` | 3.54 | 3.98 | +12.6% | 5/5 | 0 → 0 | REVIEW |
| `Rounding/DivRound` | 13.38 | 11.00 | -17.8% | 5/5 | 0 → 0 | within sample threshold |
| `Rounding/MulRound` | 13.64 | 10.73 | -21.3% | 5/5 | 0 → 0 | within sample threshold |
| `SQL/ScanInto` | 17.12 | 16.35 | -4.5% | 5/5 | 0 → 0 | within sample threshold |
| `Sqrt/FullWidth` | 139.40 | 102.40 | -26.5% | 5/5 | 0 → 0 | within sample threshold |
| `String/Canonical` | 36.64 | 37.18 | +1.5% | 5/5 | 16 → 16 | within sample threshold |
| `String/Fixed` | 35.74 | 32.04 | -10.4% | 5/5 | 16 → 16 | within sample threshold |

Review cases (1): `Rescale/Truncate`

This does **not** prove equivalence for arbitrary operands or engines.
Use Go 1.27.1 and the official v1.4.0 module before release approval.
