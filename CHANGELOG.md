# Changelog

This file describes changes present in source snapshots; it does **not**
announce that a version has been published on GitHub.

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
  See [V4 notes](docs/FASTPATH_V4.md) for measurements and known regressions.


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
