# M7-F — Controlled interleaved M7-D vs M7-E evaluation

Date: 2026-11-18. **Decision: M7-D remains the production baseline. M7-E is not promoted.**

## Scope and immutable reference

M7-D already uses imported `github.com/jokruger/dec128 v1.4.0` for 128-bit arithmetic. M7-E changes one production source file: `dec128/uint128/from_upstream_string.go`, replacing a `switch` on Jokruger parser states with a compact arithmetic translation. All division paths are byte-identical. This checkpoint adds **benchmark tooling, tests and evidence only**, without accepting the parser change.

The M7-D public API remains unchanged (1,030 historical identifiers and 66 `uint128` signatures). Official M7-D verification with Go 1.27.1 passed (provided `m7d-verify-go1271.log`); the new benchmark tooling has local Go 1.23.2 verification only.

## Official evidence: separate campaigns, not paired A/B

The original Go 1.27.1 M7-D and M7-E same-build facade-versus-Jokruger benchmarks are archived in `m7f-supplied-official/` under separate release labels. Median ns/op:

| Operation | M7-D facade | M7-E facade | Change (absolute time) | M7-D overhead vs direct | M7-E overhead vs direct |
|---|---:|---:|---:|---:|---:|
| `ParseShort` | 6.884 | 6.961 | +1.1% | +11.6% | +9.2% |
| `ParseLong` | 39.280 | 40.650 | +3.5% | +3.6% | +1.9% |
| `QuoRem` | 6.635 | 6.754 | +1.8% | +2.8% | +2.5% |
| `WideQuoRem` | 9.117 | 9.196 | +0.9% | -1.9% | 0.0% |

These runs were **not interleaved**. Jokruger direct also changed between runs, so smaller normalized overhead alone does not demonstrate that M7-E is faster.

## Controlled local A/B

A new tool, `tools/release/bench-interleaved.py`, compiles both release trees independently from temporary copies, synchronizes their benchmark sources, and repeats blocks in alternating `D-E-E-D` / `E-D-D-E` order. Every invocation tests facade and direct Jokruger with the same operands. It records raw output, flat samples, metadata, paired block deltas, upstream drift and a descriptive bootstrap interval. Neither release's source is modified.

The initial local diagnostic used Go **1.23.2**, a **temporary** copy of Jokruger source (never in the distribution), eight blocks of 120 ms, and CPU affinity pinning on a shared AMD EPYC host. The full machine-readable record is in `m7f-interleaved-local/`.

| Route | M7-D median ns/op | M7-E median ns/op | Median paired change | Block wins for E | 95% block bootstrap interval |
|---|---:|---:|---:|---:|---:|
| `ParseShort` | 9.927 | 9.690 | -4.7% | 6/8 | [-5.6%, +1.7%] |
| `ParseShortMixed` | 10.315 | 9.415 | -7.5% | 8/8 | [-9.9%, -6.1%] |
| `ParseLong` | 47.375 | 47.045 | -0.9% | 5/8 | [-3.3%, +3.2%] |
| `QuoRem` | 7.559 | 7.706 | +0.1% | 4/8 | [-3.3%, +5.9%] |
| `WideQuoRem` | 11.320 | 11.580 | +2.2% | 2/8 | [-2.2%, +17.2%] |

**Interpretation:** Mixed short input parsing consistently favored E under this Go 1.23.2 setup. Plain `ParseShort` shows possible improvement but its bootstrap interval crosses zero. Division controls fluctuate despite byte-identical production code; do not mistake that fluctuation for a proven arithmetic regression. This is **not** an official Go 1.27.1 acceptance measurement.

## Official Go 1.27.1 interleaved replication

The tool accepts the two ZIPs directly; no manual patches, module replaces or source-tree mutations are required. From the root of this M7-F checkpoint:

```sh
GOWORK=off python3 tools/release/bench-interleaved.py \
  --baseline . \
  --candidate /path/to/badassitron_M7E_tres_rutas_20261118.zip \
  --output reports/migration/m7f-official-interleaved \
  --blocks 8 --benchtime 250ms
```

It checks source isolation (`from_upstream_string.go` only), equality of `go.mod` and `go.sum`, identical benchmark code, package tests, allocations, and sample completeness. The output includes `comparison.md`, `samples.csv`, `raw.txt`, `environment.txt`, and package/compilation logs. Run `GOWORK=off bash tools/release/verify.sh` separately for the stable checkpoint; the official Go 1.27.1 technical gate has already passed for the unchanged M7-D arithmetic.

A parser improvement is only eligible for promotion when paired Go 1.27.1 measurements show a median improvement of at least 3%, an entirely negative descriptive 95% bootstrap interval, no observed control regression above 3%, and no new allocations or failed tests. Reaching the 5% *facade/direct* target is a separate criterion; M7-E's separate-campaign official overhead was still +9.2%.

## Release constraints

- M7-D remains the accepted technical release baseline; M7-E stays an unpromoted experiment.
- No arithmetic logic, public API, `go.mod`, or `go.sum` changed in M7-F.
- The shipped repository still imports Jokruger; no upstream implementation, historical engine, `vendor`, or `replace` is included.
- The historical licensing/provenance audit remains open. This checkpoint is **not** permission to publish.

## New checkpoint validation (Go 1.23.2 diagnostic toolchain)

The M7-F checkpoint was built from a fresh M7-D ZIP, then given only new benchmark/test tooling and documentation. A SHA-256 comparison of every **78 non-test Go production files** against M7-D found **zero differences**. `go.mod` and `go.sum` are also identical.

On a **separate temporary testing copy** with Go 1.23.2 and a temporary local dependency mapping (not in this repository), the following all passed: `go test ./... -count=1`, `go vet ./...`, the **entire** `go test -race -p 2 ./... -count=1` suite (exit 0), external Go client, 25 Python tool tests, backend guard and API guard. Frozen decimal oracles produced **3,518 + 1,725 = 5,243** matching results with zero unexpected differences, and independent exact-result checks reported no mismatches. Documentation link check: **384 valid local links** across 100 Markdown files (prior to appending this verification paragraph). Relevant Go test/race logs are retained in `m7f-local-verification/`.

This is **additional local verification** of unchanged M7-D production arithmetic and the added benchmarks. The **official Go 1.27.1 release-gate result belongs to M7-D**; M7-F’s benchmark tooling still needs its own official interleaved run on the target machine. No claim is made that Go 1.23.2 performance predicts Go 1.27.1 performance.
