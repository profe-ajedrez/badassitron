# Interleaved M7-D vs M7-E — parser adapter

- Toolchain: `go version go1.23.2 linux/amd64`; **3 ABBA/BAAB blocks**, 6 independent invocations per release.
- Benchmark duration: `90ms` for each subbenchmark; both binaries compiled once from copies.
- Identical benchmark sources and operands in both builds; direct Jokruger is measured in each invocation.
- The only differing production source file is `dec128/uint128/from_upstream_string.go`.
- Negative facade delta means M7-E faster. Bootstrap interval resamples blocks (descriptive, not a formal guarantee).

| Operation | M7-D facade | M7-E facade | Paired delta | 95% bootstrap CI | Direct Jokruger drift | Normalized delta | Blocks E wins |
|---|---:|---:|---:|---:|---:|---:|---:|
| `ParseShort` | 10.030 | 10.200 | +2.6% | [-2.1%, +6.4%] | +9.9% | -3.2% | 1/3 |
| `ParseShortMixed` | 10.590 | 10.380 | +3.9% | [-9.0%, +4.4%] | +3.0% | -3.5% | 1/3 |
| `ParseLong` | 48.765 | 47.450 | -3.6% | [-5.8%, +1.9%] | -0.4% | -5.4% | 2/3 |
| `QuoRem` | 7.818 | 7.883 | -0.4% | [-4.2%, +4.9%] | -3.4% | +2.1% | 2/3 |
| `WideQuoRem` | 12.010 | 11.400 | -5.6% | [-9.3%, -1.0%] | -7.8% | -1.3% | 3/3 |

**Decision gate:** NO PROMOTION: negative parser results or a control regression.

This run does not alter the M7-D release baseline. A candidate is not accepted merely because
its facade/upstream percentage improved in separate, non-interleaved campaigns.

## Block-level facade changes (%)

| Block | `ParseShort` | `ParseShortMixed` | `ParseLong` | `QuoRem` | `WideQuoRem` |
|---|---:|---:|---:|---:|---:|
| 1 | -2.11% | -8.97% | +1.95% | -0.39% | -0.99% |
| 2 | +6.35% | +4.43% | -5.84% | +4.89% | -9.31% |
| 3 | +2.64% | +3.94% | -3.60% | -4.24% | -5.65% |

## Conditions and limitations

CPU scheduling, thermal state, and code-layout differences can affect sub-nanosecond comparisons.
Direct Jokruger samples track common environmental drift; they do not remove all confounding.
Re-run with Go 1.27.1, BENCHTIME at least 250ms and at least eight blocks on the target hardware.
Functional `tools/release/verify.sh` and license/provenance approval are separate gates.

