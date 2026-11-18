# Interleaved M7-D vs M7-E — parser adapter

- Toolchain: `go version go1.23.2 linux/amd64`; **3 ABBA/BAAB blocks**, 6 independent invocations per release.
- Benchmark duration: `90ms` for each subbenchmark; both binaries compiled once from copies.
- Identical benchmark sources and operands in both builds; direct Jokruger is measured in each invocation.
- The only differing production source file is `dec128/uint128/from_upstream_string.go`.
- Negative facade delta means M7-E faster. Bootstrap interval resamples blocks (descriptive, not a formal guarantee).

| Operation | M7-D facade | M7-E facade | Paired delta | 95% bootstrap CI | Direct Jokruger drift | Normalized delta | Blocks E wins |
|---|---:|---:|---:|---:|---:|---:|---:|
| `ParseShort` | 10.315 | 9.340 | -8.1% | [-11.5%, -6.7%] | -1.1% | -7.0% | 3/3 |
| `ParseShortMixed` | 10.545 | 9.372 | -10.0% | [-19.2%, -8.8%] | -2.2% | -9.0% | 3/3 |
| `ParseLong` | 48.505 | 47.415 | -4.1% | [-8.2%, +0.1%] | -0.7% | -3.8% | 2/3 |
| `QuoRem` | 7.821 | 7.627 | -1.4% | [-4.4%, +2.5%] | +2.6% | -3.9% | 2/3 |
| `WideQuoRem` | 11.585 | 11.275 | -1.8% | [-5.1%, +25.2%] | -0.3% | -0.8% | 2/3 |

**Decision gate:** DIAGNOSTIC SIGNAL ONLY: insufficient Go version, repetitions, or benchmark duration; NO PROMOTION.

This run does not alter the M7-D release baseline. A candidate is not accepted merely because
its facade/upstream percentage improved in separate, non-interleaved campaigns.

## Block-level facade changes (%)

| Block | `ParseShort` | `ParseShortMixed` | `ParseLong` | `QuoRem` | `WideQuoRem` |
|---|---:|---:|---:|---:|---:|
| 1 | -11.51% | -19.19% | -8.17% | -1.39% | +25.24% |
| 2 | -8.09% | -8.85% | -4.14% | +2.52% | -5.08% |
| 3 | -6.71% | -9.97% | +0.15% | -4.40% | -1.75% |

## Conditions and limitations

CPU scheduling, thermal state, and code-layout differences can affect sub-nanosecond comparisons.
Direct Jokruger samples track common environmental drift; they do not remove all confounding.
Re-run with Go 1.27.1, BENCHTIME at least 250ms and at least eight blocks on the target hardware.
Functional `tools/release/verify.sh` and license/provenance approval are separate gates.

