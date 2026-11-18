# Interleaved M7-D vs M7-E — parser adapter

- Toolchain: `go version go1.23.2 linux/amd64`; **8 ABBA/BAAB blocks**, 16 independent invocations per release.
- Benchmark duration: `120ms` for each subbenchmark; both binaries compiled once from copies.
- Identical benchmark sources and operands in both builds; direct Jokruger is measured in each invocation.
- The only differing production source file is `dec128/uint128/from_upstream_string.go`.
- Negative facade delta means M7-E faster. Bootstrap interval resamples blocks (descriptive, not a formal guarantee).

| Operation | M7-D facade | M7-E facade | Paired delta | 95% bootstrap CI | Direct Jokruger drift | Normalized delta | Blocks E wins |
|---|---:|---:|---:|---:|---:|---:|---:|
| `ParseShort` | 9.927 | 9.690 | -4.7% | [-5.6%, +1.7%] | +1.6% | -3.8% | 6/8 |
| `ParseShortMixed` | 10.315 | 9.415 | -7.5% | [-9.9%, -6.1%] | -0.3% | -8.7% | 8/8 |
| `ParseLong` | 47.375 | 47.045 | -0.9% | [-3.3%, +3.2%] | -0.9% | -0.6% | 5/8 |
| `QuoRem` | 7.559 | 7.706 | +0.1% | [-3.3%, +5.9%] | +1.8% | +0.8% | 4/8 |
| `WideQuoRem` | 11.320 | 11.580 | +2.2% | [-2.2%, +17.2%] | +0.4% | +4.0% | 2/8 |

**Decision gate:** INCONCLUSIVE: keep M7-D; collect official interleaved data.

This run does not alter the M7-D release baseline. A candidate is not accepted merely because
its facade/upstream percentage improved in separate, non-interleaved campaigns.

## Block-level facade changes (%)

| Block | `ParseShort` | `ParseShortMixed` | `ParseLong` | `QuoRem` | `WideQuoRem` |
|---|---:|---:|---:|---:|---:|
| 1 | +3.96% | -6.96% | -0.72% | -7.73% | +17.16% |
| 2 | -4.99% | -4.31% | -1.59% | -2.64% | -5.03% |
| 3 | -2.78% | -8.10% | +2.25% | +6.02% | +26.90% |
| 4 | -4.57% | -9.58% | -1.03% | +5.93% | +3.23% |
| 5 | -4.73% | -6.09% | +3.19% | +2.36% | +1.22% |
| 6 | +1.68% | -6.91% | +5.35% | +5.03% | +4.25% |
| 7 | -5.60% | -9.95% | -3.83% | -3.28% | -2.19% |
| 8 | -7.14% | -10.54% | -3.32% | -2.10% | +0.51% |

## Conditions and limitations

CPU scheduling, thermal state, and code-layout differences can affect sub-nanosecond comparisons.
Direct Jokruger samples track common environmental drift; they do not remove all confounding.
Re-run with Go 1.27.1, BENCHTIME at least 250ms and at least eight blocks on the target hardware.
Functional `tools/release/verify.sh` and license/provenance approval are separate gates.

