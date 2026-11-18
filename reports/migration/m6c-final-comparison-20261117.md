# M6 public API benchmark comparison

Comparison is descriptive, not a confidence interval or release approval.
Both modes were benchmarked on the same host and toolchain, using the same inputs.

Environment: goversion=go version go1.23.2 linux/amd64, goos=linux, goarch=amd64, cpu=AMD EPYC 9V74 80-Core Processor

Provisional review trigger: greater than +10% (M2 slower).
Required samples per operation and variant: 5.

| Public operation | Default median ns/op | M2 median ns/op | Delta | Samples | B/op (default → M2) | Review |
|---|---:|---:|---:|---:|---:|---|
| `Add/MixedScale` | 10.74 | 10.26 | -4.5% | 5/5 | 0 → 0 | within sample threshold |
| `Add/NearOverflow` | 18.70 | 11.18 | -40.2% | 5/5 | 0 → 0 | within sample threshold |
| `Add/OppositeSign` | 3.51 | 4.11 | +17.2% | 5/5 | 0 → 0 | REVIEW |
| `Add/Small` | 3.68 | 3.69 | +0.1% | 5/5 | 0 → 0 | within sample threshold |
| `Compare/MixedScale` | 7.08 | 7.52 | +6.2% | 5/5 | 0 → 0 | within sample threshold |
| `Div/Exact` | 10.23 | 11.68 | +14.2% | 5/5 | 0 → 0 | REVIEW |
| `Div/Repeating` | 10.51 | 11.25 | +7.0% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/Large` | 21.99 | 16.31 | -25.8% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/ScaleOverflow` | 18.66 | 14.05 | -24.7% | 5/5 | 0 → 0 | within sample threshold |
| `Mul/Small` | 3.91 | 4.13 | +5.6% | 5/5 | 0 → 0 | within sample threshold |
| `Parse/Long` | 81.21 | 46.71 | -42.5% | 5/5 | 0 → 0 | within sample threshold |
| `Parse/Short` | 10.27 | 9.04 | -12.0% | 5/5 | 0 → 0 | within sample threshold |
| `Rescale/RoundHalfEven` | 10.12 | 5.93 | -41.4% | 5/5 | 0 → 0 | within sample threshold |
| `Rescale/Truncate` | 3.78 | 3.90 | +3.1% | 5/5 | 0 → 0 | within sample threshold |
| `Rounding/DivRound` | 13.57 | 16.19 | +19.3% | 5/5 | 0 → 0 | REVIEW |
| `Rounding/MulRound` | 14.26 | 12.35 | -13.4% | 5/5 | 0 → 0 | within sample threshold |
| `SQL/ScanInto` | 23.64 | 19.67 | -16.8% | 5/5 | 0 → 0 | within sample threshold |
| `Sqrt/FullWidth` | 173.50 | 123.30 | -28.9% | 5/5 | 0 → 0 | within sample threshold |
| `String/Canonical` | 41.32 | 44.85 | +8.5% | 5/5 | 16 → 16 | within sample threshold |
| `String/Fixed` | 41.30 | 40.18 | -2.7% | 5/5 | 16 → 16 | within sample threshold |

Review cases (3): `Add/OppositeSign`, `Div/Exact`, `Rounding/DivRound`

This does **not** prove equivalence for arbitrary operands or engines.
Use Go 1.27.1 and the official v1.4.0 module before release approval.
