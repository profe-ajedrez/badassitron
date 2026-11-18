# M7-A — Single imported 128-bit backend; unreleased source candidate

Date: 2026-11-17. Direct input: supplied cumulative M3-F ZIP; **not** an
unreviewed current GitHub checkout. Source is not published or tagged.

## Result and boundary

- Public `dec128` and public `dec128/uint128` use **imported** Jokruger v1.4.0.
- **19 historical arithmetic or rollback-only test Go files** were deleted,
  including the 12 old `dec128` arithmetic/source files, the five old
  `uint128` arithmetic/codec files and two legacy-only tests.
- Eleven active Go files had obsolete legacy-selection constraints removed.
- Unsupported rollback build tags `badassitron_legacy` and
  `badassitron_uint128_legacy` deliberately fail compilation. Alias tags
  `badassitron_m2` and `badassitron_m3` no longer select alternate engines.
- The archived original M3-F ZIP remains the recovery/provenance reference;
  its SHA-256 is in [the frozen source-API record](../../testdata/migration/api-surface/SOURCE.md).
  This ZIP is **not embedded** into the release candidate.
- The current normal tree ships **no Jokruger Go source**, `vendor`, workspace,
  or `replace`; module requirement remains Go 1.27 and Jokruger v1.4.0 with
  unchanged `go.sum`.
- `dec256`, `dec512`, and `optimus` still implement Badassitron-specific
  extended-width arithmetic and representation; this does not claim that
  Jokruger supplies their entire 256/512-bit engine.

## API and correctness evidence (local-only for this M7-A tree)

| Check | Observation |
|---|---|
| Exported historical source identifiers | **1,030/1,030** present, two additive methods |
| Exact `uint128` declaration signatures | **66/66**, unchanged |
| Default `go test ./... -count=1` | **PASS** with local Go 1.23.2 testing copy |
| `go vet ./...` | **PASS** with local Go 1.23.2 testing copy |
| M3 architecture source guard / release API and backend guard | **PASS** |
| New Python release tool tests | **16 PASS**, including strict official module-license audit controls |
| Frozen primary decimal oracle | **3,518/3,518**, zero unexpected differences |
| Frozen extended decimal oracle | **1,725/1,725**, zero unexpected differences |
| Independent exact arithmetic checks | **389 + 1,634**, zero mismatches in checked cases (skips noted in tool output) |
| Global `go test -race -p=2 ./...` | **INCOMPLETE**: local execution limit, no failed test in completed packages |
| Full core/wide package-group `-race` | **INCOMPLETE**: hit execution limit during wide `uint512` packages |
| Selected `-race`: `dec128`, `uint128`, `dec128/tests/unit`, `dec256`, `dec512` | **PASS** |
| Opt-in compatibility aliases `badassitron_m2,badassitron_m3` | **PASS** (local Go 1.23.2 copy) |
| External Go client using temporary local module replacements | **PASS** (local Go 1.23.2 copy) |
| Removed rollback tags | **FAIL CLOSED**, intentionally produce compilation error |
| Official Go 1.27.1 tests on **this M7-A ZIP** | **PENDING** |

Because this container provides Go **1.23.2**, local compile/tests used a
**separate throwaway copy** with `go 1.23`, a temporary `replace` to the
user-supplied Jokruger source snapshot, and an adapted copy of upstream's
`go.mod`. Those temporary changes are **not** in the delivered ZIP. The
supplied snapshot is not certified to be byte-identical to published v1.4.0.
The official host must rerun `tools/release/verify.sh` on M7-A using the
published module before technical release acceptance. A preliminary external
client attempt against the official Go proxy on this offline host was blocked
by DNS/network availability; the repeat with both Badassitron and Jokruger
replaced **only inside a throwaway testing module** passed. The release tree
was not changed.

The user's official Go 1.27.1 logs for the **preceding M3-F tree** are
preserved at [M3-F verification evidence](m7a-prior-m3f-host-evidence/),
including passing M2/M3 functional suites, full race checks, external fixture
and frozen oracle outputs. They do **not** certify this edited M7-A tree.

## Efficiency guard

The new [same-build benchmark](../../tools/release/README.md) compares the
public Badassitron wrapper directly against the imported Jokruger package
in one compilation, without a historical build. A **two-sample 60 ms** Go
1.23.2 diagnostic ran successfully and reported **zero heap allocations** in
all fourteen observed arithmetic and codec routes. It found, for example,
`Add` about +54% slower than direct Jokruger, `MulExact` about +19% slower,
while other cases had mixed direction. These short noisy values are not
accepted official release medians and are not evidence of end-to-end regression.
Earlier M3-F Go 1.27.1 consumer measurements showed a maximum +2.4% in its
12 sampled `dec256`/`dec512`/`optimus` operations; the present cleanup has
not been independently rebenchmarked on that host.

## Simplification and CI

- Kept the independent frozen oracles and AST API inventory, replaced live
  historical-build comparison with static source snapshots.
- Pruned migration-era benchmarking scripts, selector checkers, standalone
  candidate tooling and obsolete GitHub Actions workflows; their context and
  outputs survive in the independently held M3-F archive and existing reports.
- Replaced the two experimental CI workflows with
  `.github/workflows/release-candidate.yml` (technical gate + separate
  same-build benchmark). No publication job, version tag or registry push.
- Updated README, CHANGELOG, notices and tool entry points. The original
  M3-F ZIP remains recoverable for rollback/chronology review.

## Release blockers and decision

**Technical release check is NOT yet approved.** Need a complete Go 1.27.1+
`tools/release/verify.sh` run on the delivered source, official module metadata
comparison and reproducible benchmarks. The incomplete local full-race checks
must not be reported as PASS.

**Provenance remains blocking.** The original ownership/attribution of
historical Badassitron arithmetic sources and all required third-party notices
has not been independently verified. An MIT header or Go `go mod verify` does
not establish that ownership. No release is authorized until an accountable
maintainer completes this review.

**Decision M7-A:** ship the stripped source **as a review candidate only**,
with a single imported 128-bit engine, static compatibility baselines and an
unambiguous release gate. Do not restore the historical engine merely to
retain obsolete verification commands.
