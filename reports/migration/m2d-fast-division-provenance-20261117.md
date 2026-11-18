# M2-D — DivRound fast paths, independent rational oracle, official provenance automation

**Date:** 2026-11-17. **Source:** cumulative `badassitron_M2C_M6_procedencia_20261117.zip` (SHA-256 `f1aecf9066248c175632f4273036134a94a377768aa02f4db9b59eb1ad3784bc`). **Status:** opt-in M2 candidate, **NOT** default cutover or release approval. All source and active implementation documentation in English.

## Implemented

1. `internal/dec128candidate/round.go` and exact build-tagged mirror `dec128/m2_round.go`: the explicit `DivRoundWithStatus` path uses 128/64 or 128/128 imported Jokruger integer division when aligned operands fit in 128 bits; it keeps imported `QuoRem256By128` for a true wide numerator. Negative scale alignment retains the existing full 256-bit divisor comparison where needed. For half-rounding with a 128-bit denominator, compare `remainder` against `denominator-remainder`, avoiding a full-width multiply without losing ties. Error precedence, modes, signs and the `inexact` flag are preserved. No historical arithmetic is imported in this candidate.
2. `internal/dec128candidate/divround_exact_test.go`: independent deterministic `math/big` quotient/remainder rounding oracle, with 1,507 input pairs × six requested precisions × seven rounding modes (~63,294 evaluations). Covers large/small coefficients, positive/negative scales, all modes, signed operands, half ties, 128-bit overflows and exactly representable quotients. No use of historical or upstream decimal arithmetic to derive expected numerical answers. For mathematical overflow, the test checks the NaN result instead of asserting a boolean status on an invalid result.
3. `tools/migration-m2-cutover/audit-official.py`: checks module path/version, Go download `h1:` checksums against `go.sum`, and SHA-256 of installed official `LICENSE`/`NOTICE` against precautionary snapshot texts. Rejects a missing/unexpected or repository-local upstream directory. Reports differences instead of declaring license obligations met. Includes three offline fixture tests (six total tests of the migration tooling).
4. `verify.sh` calls the audit after official Go Modules download/verification. The migration CI workflow uploads the metadata report and module sums. README and cutover instructions now distinguish this audit from human-authored provenance review.

**No public signatures changed.** The legacy default engine, `dec256`, `dec512`, and `optimus` are not modified. The whole-package `badassitron_m2` build selection remains atomic. There is no Jokruger code copy and no `vendor/` tree. Only project-owned adaptation code uses the imported upstream integer primitives.

## Validated locally on M2-D (not official release validation)

- Platform: Linux/amd64; AMD EPYC 9V74; Go **1.23.2**.
- Local compilation used a **separate temporary modfile** with a `replace` to the user-provided **unverified** upstream snapshot. Neither that modfile nor the temporary Jokruger directory is part of this source release. The shipped `go.mod` retains `go 1.27`, `github.com/jokruger/dec128 v1.4.0`, no `replace`; `go.sum` is empty because official Go Modules resolution was unavailable.
- Default and `-tags badassitron_m2` full `go test ./... -count=1`: **PASS**. `go vet ./...` in both modes: **PASS**. Full `go test -race ./... -count=1` in both modes including `optimus`: **PASS**. Mirror/build-selection scripts: **PASS** (12 legacy vs six tagged implementation files, never mixed). Six migration-tool unit tests: **PASS**. Documentation checker: **271 local links PASS**. External Go client fixture, both modes: **PASS with temporary local dependencies**.
- Frozen primary differential corpus: **3,518 / 3,518**, unexpected differences **0**; **389** independent exact arithmetic checks, mismatches **0**. Frozen supplemental corpus: **1,725 / 1,725**, unexpected differences **0**; **1,634** independent exact checks, mismatches **0**. Skipped nonexact/invalid cases and 38 lossy encodes are not counted as exact validations. The new ~63k independent rational evaluations are separate from the frozen oracle corpus.

### Local performance evidence

A preliminary five-sample, 100ms comparison on the untouched M2-C (before this patch) flagged `Rounding/DivRound` (+19.2%), while most others were within the provisional +10% review threshold. The modified M2-D completed a **five-sample, 75ms** paired comparison of the same 20 public API operations. No reviewed regressions were flagged in that short run. Selected M2-D medians:

| Case | Default ns/op | M2-D ns/op | M2-D vs default |
|---|---:|---:|---:|
| `Rounding/DivRound` | 14.62 | 14.74 | +0.8% |
| `Mul/Large` | 23.34 | 15.59 | -33.2% |
| `Parse/Long` | 84.41 | 48.13 | -43.0% |
| `Rescale/RoundHalfEven` | 10.53 | 5.44 | -48.3% |
| `Sqrt/FullWidth` | 175.40 | 144.20 | -17.8% |
| `String/Fixed` | 43.02 | 43.01 | approximately unchanged |

All measured arithmetic operations in this suite allocated zero heap objects. A longer 200ms/five-sample collection was **interrupted partway through its tagged run** and is excluded from the comparison; the completed short-run pair and the standalone DivRound reruns are the only new M2-D measurements reported here. These short runs are **descriptive**, not sufficient to conclude release-wide speed superiority or statistically meaningful margins. Full [20-case comparison](m2d-bench-comparison-20261117.md); raw samples: [default](m2d-default-bench-20261117.txt), [tagged](m2d-tagged-bench-20261117.txt).

### Reproduction

Official host (with Go 1.27.1 and official v1.4.0 available):

```bash
GOWORK=off bash tools/migration-m2-cutover/verify.sh
BENCHTIME=200ms COUNT=5 bash tools/migration-m2-cutover/bench-m6.sh reports/migration/m2d-official-bench
```

The first command generates `go.sum` and an `m2d-official-module-audit.md` metadata report when the official download succeeds; **commit the verified Go checksums before proposing a release**. These steps were **not run successfully in the present offline container**.

## Provenance and release gates

The version-specific Go package listing of `github.com/jokruger/dec128@v1.4.0` describes MIT and attributes code to Quang (BSD-3-Clause, 2025) and Luke Champine (MIT, 2019): <https://pkg.go.dev/github.com/jokruger/dec128@v1.4.0>. This external listing supports the stated **license class and attribution names**, not identical license-file bytes or a complete source-chain audit. The supplied snapshot's texts remain precautionary; the new tool can compare them against the actual downloaded tag. Historical Badassitron source ownership and legal notice sufficiency remain human-review items.

| Stage | M2-D checkpoint |
|---|---|
| M0 | Partial — official hashes/notices and historical source chain not closed |
| M1 | Bridge and independent codecs verified locally; prior M2-B official Go 1.27 log is historical evidence |
| M2 | Whole-public-package opt-in candidate passes local suites/oracles/race; default is legacy |
| M3 | Not started — public uint128 migration deferred until atomic M2 approval |
| M4–M5 | Consumer regressions pass in both build modes; no source migration |
| M6 | 20-case local matrix and exact-divide optimization complete; official 1.27.1/real workloads pending |
| M7 | **Blocked** — do not publish or change default until Go 1.27.1/M2-D, checksums, M0 and representative M6 sign-off |

Historical reports are retained unchanged; no performance figures, attestations or upstream module checksums are invented.
