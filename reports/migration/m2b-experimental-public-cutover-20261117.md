# M2-B: opt-in atomic public cutover and performance hardening

**Date:** 2026-11-17. **Classification:** experimental, not a release.
**Input:** `badassitron_M1_final_M2_candidate_A_20261117.zip` (the
cumulative, locally validated M2-A candidate), unpacked as the working
baseline. No GitHub branch or old source checkout replaced that baseline.

## Implemented

1. Staged a **whole-package opt-in replacement**: 12 historical public
   implementation files are compiled only without the `badassitron_m2`
   build tag, and six tagged `dec128/m2_*.go` source files compile the
   Jokruger-based M2 candidate under the **same nominal public `Dec128` type
   and import path**. No legacy arithmetic source is included in the tagged
   `dec128` build. The default build remains unchanged in its arithmetic.
2. Kept `internal/dec128candidate` as an independent comparison target.
   Its code is temporarily mirrored in the six tagged public files until
   official verification permits retiring the old engine. A new
   `tools/migration-m2-cutover/check-mirror.py` enforces source parity.
3. Optimized candidate/public-M2 Add/Sub on differing scales using imported
   Jokruger `uint128` exact alignment, carrying the historical canonical
   retry, sign and signed-zero semantics. This removes fallback dependence
   on Jokruger's global arithmetic policy and the expensive conversion to
   `jk.Dec128` for those cases.
4. Reworked exact `Rescale` to use upstream unsigned multiplication/division
   and preserve stored negative-zero metadata, without calling historical
   code. Specialized `RescaleRoundWithStatus` for a downscale at most 19
   decimal places, using the exact `uint64` remainder and all seven public
   rounding modes; it preserves the distinction between *already-negative
   zero* and *negative nonzero rounded to zero*.
5. Added `testdata/migration/external-client`, a Go client importing only
   public package paths. It compiles and tests Go interface method sets,
   generic byte/string constructors, signed zero, NaN propagation, SQL,
   JSON, and the 256/512/Optimus consumer surface. Also tested a genuine
   *separate-module* copy with a temporary local replacement strictly
   outside the delivered repository.
6. Added 20-operation public-method microbenchmark matrix, release-gate
   scripts in `tools/migration-m2-cutover`, and a migration-branch/manual
   GitHub Actions workflow for Go 1.27 and official Jokruger v1.4.0.
   **The workflow is not yet executed.**
7. Converted two historical white-box signed-zero tests to assertions against
   public `StoredNegative`, so they run in both builds. Two implementation-
   specific tests remain explicitly legacy-only because they need removed
   private helpers/state. Their public behavior is additionally covered by
   the golden oracles and candidate tests.

These are staged source files, not the final elimination of the old engine.
The target still has a temporarily duplicated candidate for inspection and
independent testing. The release plan must remove that duplication after the
cutover gate is approved, not conceal it.

The build-selection audit explicitly observed **12 default legacy source files**
versus **6 tagged M2 source files**, with no arithmetic source mixing.

## Verified locally

**Environment:** `linux/amd64`, AMD EPYC 9V74, Go **1.23.2**.
A temporary `-modfile` at `/mnt/data/session_M2B/local.mod` changed the
minimum Go version only for **testing** and referenced a separate upstream
`dec128-main` snapshot. That snapshot is not verified as the v1.4.0 tag.
Neither file is included in this distribution. The shipped `go.mod`
retains Go **1.27** and `github.com/jokruger/dec128 v1.4.0` with **no**
`replace`; the official `go.sum` is still empty.

| Gate | Result | Evidence |
|---|---|---|
| `check-mirror.py` / `check-build-selection.py` | **PASS** — six mirrors match, mutually exclusive source selection | Script output and source inspection |
| `go test ./... -count=1` default | **PASS** | [default suite](m2b-default-tests-20261117.txt) |
| `go test -tags badassitron_m2 ./... -count=1` | **PASS**, including `optimus`, `dec256`, `dec512` | [tagged suite](m2b-cutover-tests-20261117.txt) |
| `go vet ./...` and `go vet -tags badassitron_m2 ./...` | **PASS** | [tagged vet, empty output](m2b-cutover-vet-20261117.txt) |
| M2 **public** 3,518-case primary oracle | **PASS**, 0 unexpected differences | [comparison](m2b-public-primary-compare-20261117.txt) |
| M2 **public** 1,725-case supplemental oracle | **PASS**, 0 unexpected differences | [comparison](m2b-public-extended-compare-20261117.txt) |
| Independent mathematical checks | **PASS**: 389 primary + 1,634 extended, zero mismatches; 38 lossy encodings explicitly counted | [primary](m2b-public-primary-exact-20261117.txt), [extended](m2b-public-extended-exact-20261117.txt) |
| Candidate seeded randomized arithmetic, rounding and sqrt tests | **PASS** after optimized arithmetic changes | Included in tagged/default suites |
| Separate third-party Go module, both tag modes | **PASS** (local Go 1.23 override only) | [default](m2b-external-default-20261117.txt), [tagged](m2b-external-tagged-20261117.txt) |
| `-race` for public/wide packages, bridge, candidate, external-client | **PASS** scoped | [scoped race](m2b-cutover-race-scoped-20261117.txt) |
| `-race` selected Optimus tests | **PASS** selected subset | [Optimus subset](m2b-cutover-optimus-race-subset-20261117.txt) |
| `-race` **entire** repository under M2 tag | **INCOMPLETE**; stopped by sandbox execution limit before `optimus` finished | [partial log](m2b-cutover-race-incomplete-20261117.txt) |
| Differential tool tests | **PASS**, 6 Python tests | [output](m2b-oracle-tool-tests-20261117.txt) |
| Documentation link checker | **PASS**: 210 links in 56 files before this report was added | [output](m2b-doclinks-20261117.txt) |
| Official Go 1.27 and published Jokruger v1.4.0 | **BLOCKED**: toolchain and network/module checksums unavailable locally | Official CI procedure remains unexecuted |

The oracle captures serialized text/JSON/SQL fields, stored scale, error
reasons, and rounding status, but it cannot prove full compatibility for all
128-bit operands, arbitrary third-party drivers or binary ABI. The public
M2 variant is still **not the default** and is **not release-certified**.

## M6 performance evidence: 20 public operations

Both public variants were measured on this host with Go 1.23.2,
`GOMAXPROCS=1`, five 80ms samples, `-benchmem`. The raw files are
[default](m2b-bench-default-20261117.txt) and
[tagged before the final round-specific optimization](m2b-bench-tagged-round-before-20261117.txt).
Medians are nanoseconds per operation. The final rounding specialization was
measured separately at 100ms × 5: [rounding after](m2b-bench-round-after-20261117.txt).
All sampled arithmetic operations have **zero allocations**; String
allocates once in both builds.

| Operation | Default ns/op | Tagged M2 ns/op | Difference |
|---|---:|---:|---:|
| Add, same scale | 3.57 | 3.62 | +1.3% |
| Add, mixed scale | 10.60 | 11.33 | +6.9% |
| Add, near overflow | 18.84 | 11.37 | -39.6% |
| Compare, mixed scale | 7.27 | 7.27 | 0% |
| Mul, small | 4.18 | 4.07 | -2.7% |
| Mul, large | 22.15 | 25.62 | +15.7% |
| MulRound | 16.05 | 12.75 | -20.6% |
| Div, exact | 10.01 | 11.00 | +9.9% |
| DivRound | 14.98 | 16.32 | +8.9% |
| Parse, long | 79.49 | 44.05 | -44.6% |
| Parse, short | 10.00 | 10.73 | +7.3% |
| Rescale, truncate | 3.90 | 4.25 | +9.0% |
| RescaleRound, half-even (pre-final optimization) | 10.29 | 12.76 | +24.0% |
| RescaleRound, half-even (separate final five-sample run) | 10.29 (earlier) | **5.64** | approx. -45%, not a paired benchmark |
| ScanInto from string | 24.77 | 28.52 | +15.1% |
| Sqrt, full-width | 189.30 | 224.10 | +18.4% |

The first experiment exposed large regressions (Add with mixed scales
+113%, near overflow +132%, truncate +128%, half-even rescale +79%).
Those paths were optimized rather than accepted. Some remaining sampled
regressions are still above the provisional 10% performance budget, notably
large multiplication, SQL scanning and sqrt. Do not claim M6 is GREEN.
`benchstat` is unavailable here; these are descriptive medians, not
statistically established universal regressions. They cannot substitute for
same-host official Go 1.27 / v1.4.0 measurements and full workloads.

## Gate decision and stage status

**M2 is still gated**: no default cutover, no upstream code vendored,
no historic code removed from the default public build, no new tag published.
The opt-in tagged package validates the actual integration route, but
final promotion requires official tag verification, checksums, Go 1.27,
external consumer tests, complete global race run and accepted M6 budget.

| Stage | State after M2-B |
|---|---|
| M0 | **Partial** — attribution/provenance review and official source checksum unfinished |
| M1 | **Advanced, partial** — independently tested bridge and codecs, but Go 1.27/official module gates open |
| M2 | **Experimental whole-public-package cutover implemented and locally verified**; default remains historical |
| M3 | **Pending**; public `uint128` not substituted (dependent on approved M2) |
| M4 | **Regression-tested only**; `dec256`/`dec512` no migration yet |
| M5 | **Regression-tested only**; `optimus` no tier migration yet |
| M6 | **Partial**; expanded 20-case matrix, several operations exceed provisional budget |
| M7 | **Blocked** on M0–M6, official module/checksum and supported toolchain |

## How to reproduce in an authorized environment

```sh
# Go 1.27; no local replacement; working upstream registry connectivity.
bash tools/migration-m2-cutover/verify.sh

# Focused public-path benchmarks, both variants on one machine:
GOMAXPROCS=1 GOWORK=off go test ./dec128/tests/benchmark \
    -run '^$' -bench '^BenchmarkM2Cutover$' -benchtime=200ms -count=5 -benchmem
GOMAXPROCS=1 GOWORK=off go test -tags badassitron_m2 ./dec128/tests/benchmark \
    -run '^$' -bench '^BenchmarkM2Cutover$' -benchtime=200ms -count=5 -benchmem
```

The GitHub Actions workflow is `.github/workflows/m2-experimental-cutover.yml`.
An official `go.sum` must be committed after validating its origin; this
checkpoint does not fabricate a checksum or a successful online build.


### AST inventory caveat

The `tools/migration-api` source inventory now respects Go build tags via
`-tags badassitron_m2` rather than accidentally including both engines in a
single parse. Compared with the M0 snapshot of 1,033 declarations, the
default build contains 1,035 (only the documented additive `StoredNegative`
and `NewStoredZero`). The tagged build contains 1,037. Its raw textual
inventory reports 60 differences in declaration text, predominantly parameter
identifier names and import aliases, private `Dec128` storage, and constant
initializer spelling; one former `sqlScanDestination.Scan` method belongs to
an *unexported* implementation type replaced by another unexported scanner.
The candidate additionally exposes `ErrorCode` and `DivAtPrecision`.
Those raw AST text differences are **not a validated Go type-equivalence proof**.
External import and method-set compile fixtures passed locally, but final
Go 1.27 public API/types and third-party driver checks remain open.
