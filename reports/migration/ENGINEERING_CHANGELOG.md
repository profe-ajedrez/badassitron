# Engineering changelog (archived)

> Dated internal milestones and experiments. For user-facing changes, see the
> [public changelog](../../CHANGELOG.md) and [documentation](../../docs/README.md).

This file describes changes present in source snapshots; it does **not**
announce that a version has been published on GitHub.

## M7-H documentation integrity and repository hygiene (unpublished; 2026-11-18)

- Keep the complete M7-D production backend unchanged and do not promote
  any M7-G parser candidate. Update README, current release-status and
  provenance notices to distinguish archived source-state claims from the
  active checkout and to identify the continuing publication blocker.
- Add an explicit migration-history index and repository-hygiene controls.
  Move obsolete, unreferenced generated evidence into a separate archive;
  preserve links to in-tree evidence and a full record outside the clean ZIP.
- The exact final M7-H source still requires its own official Go 1.27.1
  verification against the published Jokruger v1.4.0 module; earlier M7-D
  evidence does not waive this requirement. See the [M7-H audit](notes/Informe_M7H_repository_cleanup.md).

## M7-G parser audit (rejected; 2026-11-18)

- Test three bounded adapter hypotheses without adding an independent parser.
- The supplied official M7-D/M7-E interleaved comparison reported a 2.1%
  `ParseShort` improvement, below the 3% required threshold. A separate
  local Go 1.23.2 benchmark against a provided Jokruger `main` snapshot
  had an inconclusive interval for its `ParseShort` improvement.
- Preserve **M7-D** parser source; functional local validation and raw
  evidence are recorded in [M7-G report](notes/Informe_M7G_parser_final.md).

## M7-F controlled A/B harness (2026-11-18)

- Keep the M7-D production code unmodified. Add parser contract tests,
  interleaved A/B tools and supplied benchmark evidence. M7-E is not
  promoted. See [M7-F report](m7f-interleaved-evaluation-20261118.md).

## M7-D imported AddCarry adapter (unpublished; 2026-11-18)

- Optimize public `Uint128.Add` using Jokruger's exported `AddCarry` primitive.
  A two-limb carry mask preserves the legacy zero-on-overflow result and error
  code, without copying an arithmetic algorithm or modifying the public API.
- Test 24,000 deterministic upstream comparisons, 4,096 independently checked
  `math/big` additions and wide-divider error precedence. Preserve the M7-C
  implementation of other arithmetic and parsing where local experiments
  threatened successful-operation inlining or failed to improve parsing.
- Add a configurable A/B comparator for M7-C versus M7-D, and retain the
  supplied official M7-C Go 1.27.1 benchmark data as an immutable baseline.
- Local Go 1.23.2 full suite, vet, 5,243 frozen oracles, and grouped race
  tests passed. The global race command timed out; **Go 1.27.1 verification
  and performance measurement of this M7-D source remain pending**.
  Historical Badassitron provenance and permission remain unresolved.
- [M7-D review](m7d-addcarry-review-20261118.md).

## M7-C imported-adapter tuning (unpublished; 2026-11-18)

- Delegate all arithmetic to the same imported Jokruger v1.4.0 backend,
  with narrower, branchless historical error translations for the pinned
  unsigned operation states. Removed duplicate checks already performed
  upstream; retained the public API and zero-on-error behavior.
- Add 12,000 deterministic compatibility comparisons, numeric state pin
  tests, a reproducible old/new direct-upstream benchmark comparator and
  preserved local evidence.
- Full local functional tests, vet, 5,243 frozen oracles, and race tests by
  package passed. The global race command timed out; Go 1.27.1 technical
  verification of M7-C is pending. Source provenance remains unresolved.
- [M7-C evidence](m7c-adapter-performance-20261118.md).

## M7-A single-backend source consolidation (unpublished; 2026-11-17)

- Removed the historic build-tagged `dec128` and `uint128` arithmetic sources
  from the release-candidate tree; default 128-bit arithmetic comes from the
  imported Jokruger v1.4.0 module.
- Froze the pre-migration 1,030 public identifiers and 66 unsigned declaration
  signatures into deterministic source-API comparison artifacts.
- Removed experimental rollback build modes; obsolete rollback tags now fail
  explicitly rather than silently compiling a different implementation.
- Added one release gate for Go 1.27+, frozen oracles, race, third-party module
  metadata, external clients and Markdown links, plus same-build direct-upstream
  benchmarks and CI.
- M7-A completed its official Go 1.27.1 verification, but source provenance
  remains unapproved. **No release or tag created.**

## M3-A unsigned-128 opt-in experiment (unpublished; 2026-11-17)

- Added `badassitron_m3` as an optional, Jokruger-backed public unsigned
  arithmetic facade; the historical `dec128/uint128` remains the default.
- Preserved all 66 public declarations and exact signatures, with explicit
  historic error translation and no type alias or copied upstream Go source.
- Added 12,000 independent seeded unsigned operand cases, complete local
  suites/races, an independent Go client, build-selection/API checks, and an
  official-host benchmark/verification workflow.
- **Rejected default M3 promotion:** a same-host Go 1.23.2 benchmark showed
  +90.3% Add and +26.1% overflowing Mul medians. Official Go 1.27.1/module
  validation and source provenance remain release gates.
- See [M3 report](m3-uint128-opt-in-20261117.md).

## M2-E unpublished source checkpoint (2026-11-17)

- Select the Jokruger-backed `dec128` implementation by default; retain
  historical arithmetic using the explicit `badassitron_legacy` build tag.
- Preserve the `badassitron_m2` alias and add an atomic file-selection check
  for both engines. No public import path is changed.
- Add independent arbitrary-precision truncation and signed-zero tests in
  both modes, adapt the full verification and benchmark scripts to the new
  default, and preserve received Go 1.27.1/M2-D evidence as historical data.
- Record the official module `h1:` sums supplied by the verified download
  log. Revalidation on the final tree and historical provenance review remain
  prerequisites for a release; no tag has been created.
- Retain a provisional M6 release review on `Rescale/Truncate` (+0.44 ns/op,
  +12.6% in a five-sample Go 1.27.1 benchmark). The unhelpful uint64
  optimization experiment was **reverted**.

## Documentation and onboarding pass (current source; unpublished)

- Added an immediately runnable README example, a ten-minute introduction,
  a decision guide for decimal behavior, and a documentation map.
- Fixed broken internal links to audit documents not shipped in this source
  archive; preserved historical audit references as names, not dead links.
- Clarified the difference between `RoundBank` and `RescaleRound`, SQL scanning,
  empty input, `NaN` and inexactness, and the currently required Go version.
- Expanded `dec128` package GoDoc and added runnable GoDoc examples.
- Added a documentation link check for future changes.
- No arithmetic algorithms or exported API declarations changed.

## Dec128 FASTPATH V4 (current source; not a release announcement)

- Optimized the common paths for `Add`, `Sub`, `Mul`, and `Compare`.
- Packed sign and all 256 possible `errors.Error` values into a `uint16`
  private state. The type remains 24 bytes on amd64; exported API unchanged.
- Verified the corrected candidate against V3 and measured it on Go 1.27.1.
  See [V4 notes](notes/FASTPATH_V4.md) for measurements and known regressions.


Older entries are retained for historical context, including their original
verification limitations. Those limitations are not the status of the current
checkout. Entries do not imply a new semantic version tag or a published release.

## Optimus v8 documentation revision

### Documentation

- Rebuilt the README around install, type selection, short examples, SQL,
  exact conversions, and verification.
- Replaced the outdated Optimus v1 stage guide with a current v8 user guide.
- Updated the historical decimal and integer stage documents to make their
  original scope clear without describing delivered packages as future work.
- Added GoDoc to the exported Optimus API and supporting decimal and integer
  APIs, with special attention to scale, rounding, errors, and SQL.
- Added a short guide for the existing staged calculation workflow.
- Added runnable GoDoc examples for rounding, inexact division, and SQL.
- Added this changelog and documented the remaining verification limits.

### Compatibility

- Documentation and GoDoc changes only.
- No public API declarations or numeric algorithms changed in this revision.

## Optimus v8. Phase C precision audit

### Fixed

- Corrected the small coefficient `QuoRem` scale alignment path. It now uses
  `uint128.MulCarry` and accepts the fast path only when the high product is
  zero. An overflowing alignment falls back to an exact calculation.
- The fix also prevents the same incorrect remainder from reaching `Mod`.

### Tests

- Added a deterministic regression for the overflowing 128 bit alignment.
- Added directed precision, rounding, comparison, conversion, SQL, and
  representation checks with independent rational oracles.
- Added differential fuzzing seeds and a targeted quotient and remainder
  benchmark.
- The v8 audit reports 628245 fuzz executions across six targets without
  another reproduced discrepancy.

### Compatibility and performance

- The exported Optimus API is unchanged from v7.
- Other packages and existing numeric algorithms remain unchanged.
- The former incorrect fast path now uses a slower exact fallback. The
  audited case rose from about 42 ns per operation to about 655 ns per
  operation on the recorded environment. Correct fast paths remain in use.

### Known limits recorded by the v8 audit (historical)

- At the time, the full repository test command was blocked by compilation
  errors in `withfloat64/internal/math.go`; those errors were later corrected.
- That audit used Go 1.23.2 with an external temporary modfile. It did not
  verify the then-declared Go 1.24 toolchain. The current checkout declares
  Go 1.27 and has its own validation requirements.
- Two inherited `uint128` issues are documented but remain outside the v8
  Optimus correction. They concern overflow reporting in `Mul` and mutation
  of the argument passed to `FromBigInt`.
- Historical `Optimus.Scan` accepts floating point input and converts NULL
  to zero. `SQLDecimal` provides a separate exact and nullable path.

## Optimus v7. Phase C concurrency

- Added concurrency checks for shared immutable Optimus values and indirect
  coefficient storage.
- Verified atomic default precision access and concurrent SQL and text use.
- See v7 audit (historical audit; not bundled).

## Optimus v6. Phase C SQL

- Added `SQLDecimal` for exact scanning and nullable SQL values.
- Retained the older `Optimus.Scan` and `Optimus.Value` behavior.
- See v6 audit (historical audit; not bundled).

## Optimus v4 and v5. Phase B

- Improved quotient, remainder, and wide multiplication fast paths, with
  exact fallback calculations where needed.
- Added more independent arithmetic and benchmark coverage.
- See v4 audit (historical audit; not bundled) and
  v5 audit (historical audit; not bundled).

## Optimus v1 to v3. Foundation and Phase A

- Added the adaptive decimal type and its 64, 128, 256, and 512 bit tiers.
- Added exact decimal conversions and explicit rounding operations.
- Improved extreme integer powers with interval based rounding checks.
- See v1 audit (historical audit; not bundled),
  v2 audit (historical audit; not bundled), and
  v3 audit (historical audit; not bundled).
