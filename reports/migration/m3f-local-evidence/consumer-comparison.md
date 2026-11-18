# M3-F representative consumer benchmarks (LOCAL, exploratory)

Go 1.23.2 / Intel Xeon Platinum 8573C; three independent 100ms samples; mixed scheduling noise. Negative = new unsigned facade faster. Identical M2 decimal engine in both builds.

| Package | Benchmark | Historical unsigned (ns/op) | Imported unsigned (ns/op) | Change | Allocations both |
|---|---|---:|---:|---:|---|
| `dec256` | `Add` | 50.65 | 53.68 | +6.0% | `(0, 0)` vs `(0, 0)` |
| `dec256` | `DivRound` | 245.80 | 214.70 | -12.7% | `(0, 0)` vs `(0, 0)` |
| `dec256` | `Mul` | 90.29 | 93.43 | +3.5% | `(0, 0)` vs `(0, 0)` |
| `dec512` | `Add` | 141.40 | 126.80 | -10.3% | `(0, 0)` vs `(0, 0)` |
| `dec512` | `DivRound` | 634.50 | 775.10 | +22.2% | `(0, 0)` vs `(0, 0)` |
| `dec512` | `Mul` | 324.40 | 338.20 | +4.3% | `(0, 0)` vs `(0, 0)` |
| `optimus` | `AddSmall` | 34.63 | 33.68 | -2.7% | `(0, 0)` vs `(0, 0)` |
| `optimus` | `AddWide` | 178.20 | 205.50 | +15.3% | `(64, 1)` vs `(64, 1)` |
| `optimus` | `DivRepeating` | 58.43 | 87.07 | +49.0% | `(0, 0)` vs `(0, 0)` |
| `optimus` | `DivSmall` | 33.94 | 51.89 | +52.9% | `(0, 0)` vs `(0, 0)` |
| `optimus` | `MulSmall` | 38.35 | 35.72 | -6.9% | `(0, 0)` vs `(0, 0)` |
| `optimus` | `MulWide` | 318.10 | 329.20 | +3.5% | `(64, 1)` vs `(64, 1)` |

These short results cannot prove a representative performance win. Repeat with Go 1.27.1 and longer alternating samples.
