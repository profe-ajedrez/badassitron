# M6 public API benchmark comparison

Comparison is descriptive, not a confidence interval or release approval.
Both modes were benchmarked on the same host and toolchain, using the same inputs.

Environment: goversion=go version go1.23.2 linux/amd64, goos=linux, goarch=amd64, cpu=AMD EPYC 9V74 80-Core Processor

Provisional review trigger: greater than +10% (default M2 slower).
Required samples per operation and variant: 5.

| Public operation | legacy median ns/op | default M2 median ns/op | Delta | Samples | B/op (legacy → default M2) | Review |
|---|---:|---:|---:|---:|---:|---|
| `Add/MixedScale` | 10.04 | 10.81 | +7.7% | 5/5 | 0 → 0 | within sample threshold |
| `Add/NearOverflow` | 18.32 | 10.75 | -41.3% | 5/5 | 0 → 0 | within sample threshold |
| `Add/OppositeSign` | 3.40 | 3.85 | +13.2% | 5/5 | 0 → 0 | REVIEW |
| `Add/Small` | 3.48 | 3.71 | +6.7% | 5/5 | 0 → 0 | within sample threshold |
| `Compare/MixedScale` | 7.01 | 6.89 | -1.8% | 5/5 | 0 → 0 | within sample threshold |
| `Div/Exact` | 9.84 | 10.28 | +4.5% | 5/5 | 0 → 0 | within sample threshold |
| `Div/Repeating` | 10.28 | 10.58 | +2.9% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/Large` | 21.77 | 15.24 | -30.0% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/ScaleOverflow` | 18.12 | 13.19 | -27.2% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/Small` | 3.91 | 3.79 | -3.3% | 5/5 | 0 → 0 | within sample threshold |
| `Parse/Long` | 76.98 | 41.61 | -45.9% | 5/5 | 0 → 0 | within sample threshold |
| `Parse/Short` | 9.80 | 8.30 | -15.3% | 5/5 | 0 → 0 | within sample threshold |
| `Rescale/RoundHalfEven` | 10.19 | 5.55 | -45.6% | 5/5 | 0 → 0 | within sample threshold |
| `Rescale/Truncate` | 3.86 | 3.91 | +1.3% | 5/5 | 0 → 0 | within sample threshold |
| `Rounding/DivRound` | 15.40 | 13.52 | -12.2% | 5/5 | 0 → 0 | within sample threshold |
| `Rounding/MulRound` | 14.82 | 12.47 | -15.9% | 5/5 | 0 → 0 | within sample threshold |
| `SQL/ScanInto` | 24.28 | 20.56 | -15.3% | 5/5 | 0 → 0 | within sample threshold |
| `Sqrt/FullWidth` | 189.90 | 128.00 | -32.6% | 5/5 | 0 → 0 | within sample threshold |
| `String/Canonical` | 40.87 | 44.92 | +9.9% | 5/5 | 16 → 16 | within sample threshold |
| `String/Fixed` | 40.37 | 39.93 | -1.1% | 5/5 | 16 → 16 | within sample threshold |

Review cases (1): `Add/OppositeSign`

This does **not** prove equivalence for arbitrary operands or engines.
Use Go 1.27.1 and the official v1.4.0 module before release approval.
