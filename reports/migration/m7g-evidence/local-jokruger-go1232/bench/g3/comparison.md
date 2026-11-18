# Interleaved M7-D vs M7-E — parser adapter

- Toolchain: `go version go1.23.2 linux/amd64`; **3 ABBA/BAAB blocks**, 6 independent invocations per release.
- Benchmark duration: `90ms` for each subbenchmark; both binaries compiled once from copies.
- Identical benchmark sources and operands in both builds; direct Jokruger is measured in each invocation.
- The only differing production source file is `dec128/uint128/from_upstream_string.go`.
- Negative facade delta means M7-E faster. Bootstrap interval resamples blocks (descriptive, not a formal guarantee).

| Operation | M7-D facade | M7-E facade | Paired delta | 95% bootstrap CI | Direct Jokruger drift | Normalized delta | Blocks E wins |
|---|---:|---:|---:|---:|---:|---:|---:|
| `ParseShort` | 9.909 | 9.970 | +1.8% | [-2.4%, +3.0%] | -1.4% | +1.6% | 1/3 |
| `ParseShortMixed` | 10.260 | 10.470 | -1.4% | [-3.7%, +4.4%] | -0.6% | -3.1% | 2/3 |
| `ParseLong` | 47.325 | 46.055 | +1.2% | [-7.2%, +1.4%] | +0.2% | +1.3% | 1/3 |
| `QuoRem` | 8.286 | 7.720 | -5.1% | [-9.9%, +0.7%] | -5.4% | +3.2% | 2/3 |
| `WideQuoRem` | 11.305 | 11.225 | -1.3% | [-14.2%, +0.3%] | -5.7% | -4.8% | 2/3 |

**Decision gate:** INCONCLUSIVE: keep M7-D; collect official interleaved data.

This run does not alter the M7-D release baseline. A candidate is not accepted merely because
its facade/upstream percentage improved in separate, non-interleaved campaigns.

## Block-level facade changes (%)

| Block | `ParseShort` | `ParseShortMixed` | `ParseLong` | `QuoRem` | `WideQuoRem` |
|---|---:|---:|---:|---:|---:|
| 1 | +1.79% | -3.72% | -7.18% | +0.75% | -1.31% |
| 2 | +3.00% | +4.38% | +1.42% | -5.07% | +0.31% |
| 3 | -2.45% | -1.38% | +1.21% | -9.86% | -14.22% |

## Conditions and limitations

CPU scheduling, thermal state, and code-layout differences can affect sub-nanosecond comparisons.
Direct Jokruger samples track common environmental drift; they do not remove all confounding.
Re-run with Go 1.27.1, BENCHTIME at least 250ms and at least eight blocks on the target hardware.
Functional `tools/release/verify.sh` and license/provenance approval are separate gates.

