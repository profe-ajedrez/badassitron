> **Historical migration record (archived at M7-H).** This file contains

> **Archived engineering record.** This document describes a dated source
> checkpoint, not the currently recommended API or release state.
> For supported usage, see the [documentation index](../../../docs/README.md).
> time-specific stage status, proposed code and, in places, old rollback
> tag instructions that **do not apply to the current source tree**.
> The authoritative state is [M7-H release status](../../../docs/RELEASE_STATUS.md);
> do not follow an old "current" or "pending" claim as a live rule.

# M0 — Decimal migration contract

## M3-F default public unsigned backend (2026-11-17)

The normal build of `dec128/uint128` now selects imported Jokruger arithmetic,
bitwise and wide division methods. Badassitron's nominal `Uint128` value and
`errors.Error` return types remain unchanged. `FromString` and `StringToBuf`
use imported codecs, with the pre-migration behavior tested against Jokruger.
Temporary historical source files require `badassitron_uint128_legacy` or
`badassitron_legacy` (unless `badassitron_m3` forces imported unsigned).
This supersedes M3-E's default-selection decision and does not change its
historical evidence. See [M3-F report](../m3f-default-imported-uint128-20261117.md).
Official Go 1.27.1 validation is still required before publication.


**Status:** draft frozen baseline; awaiting review and Go 1.27 verification.
**Source:** `badassitron-master(6).zip`, SHA-256
`d6628a2fd4852d06343b95ead5b8e0351d7994acd401921faeb38eb60b6bf7df`.
**Proposed dependency:** `github.com/jokruger/dec128 v1.4.0`, not installed.

## Scope and authority

The machine-readable declaration snapshot is
[`testdata/migration/api-baseline.json`](../../../testdata/migration/api-baseline.json).
Regenerate with `GO111MODULE=off go run ./tools/migration-api -root .`.
All exported type, function, method, variable, and constant declarations from
`dec128`, `dec128/errors`, `dec128/uint128`, `dec256`, `dec256/uint256`,
`dec512`, `dec512/uint512`, `optimus`, `withdec128`, and `withfloat64`
are included. A declaration snapshot is **not** proof of full source/API
compatibility: embedded methods, interface satisfaction, generic constraints,
field visibility, type identity, and external consumer compilation need separate
checks. The source package import paths must remain stable.

## Observed public behavior that must not silently change

1. `Dec128` has private coefficient, scale and state. Consumers may not rely on
   its memory layout. `Precision()` and `Exponent()` report stored fractional
   scale, not significant-digit count. The maximum stored scale is 19.
2. `StringFixed()` retains stored fractional zeroes and `String()` omits
   insignificant fractional zeroes. `Canonical()` can change stored scale
   without changing numeric value.
3. `Add`, `Sub`, and `Mul` aim for exact representability; failed `Mul`
   historically reports `NaN(errors.Overflow)` even for excess scale. A
   replacement that silently rounds is not source-semantic compatible.
4. `Round(prec)` is half-away-from-zero. `RoundBank` is ties-to-even.
   Existing `RoundingMode` values must retain their numeric representation.
5. The historical parser accepts empty input as zero and `"5."` as five.
   Preserve both until a documented change is approved.
6. `NaN` propagates arithmetic errors; `ErrorDetails()` returns the public
   `errors.Error` representation. All 0–10 historical `errors.Error` codes
   retain their numeric values; new codes may only append.
7. `Equal` treats two NaN values as equal, regardless of error category.
   Equality ignores scale and the sign of zero for numeric values.
8. `WithStatus` booleans mean *nonzero digits discarded*, not an error flag.
9. `SetDefaultPrecision` is library-local state; the new wrapper must not
   accidentally inherit mutable Jokruger process-wide policy.
10. `Scan` currently has a value receiver and fails to update its caller.
    `ScanInto` and `ScanDestination` are the working migration choices. A
    pointer receiver changes interface method sets and needs an explicit
    compatibility decision.
11. `dec128/uint128` is an independently exported API. Wider decimals and
    Optimus must avoid conversion overhead in their arithmetic hot loops.
12. Signed zero, NaN reasons, JSON/text/binary marshaling, driver.Value,
    and `database/sql` behavior require method-specific golden coverage.

## Acceptance criteria

- Exact exported symbol inventory matches, or every difference is approved.
- Behavioral golden fixtures for arithmetic, parse, zero, NaN, serialization,
  rounding, SQL, and boundaries are checked against the old and new versions.
- Dependency globals are not used as hidden operational inputs.
- Per-method contextual error mapping is tested, not merely ordinal mapping.
- Decimal widths 128/256/512 and Optimus remain capable of their current
  scale and range, independently of the upstream 128-bit limit.
- Differential tests compare against a **separately preserved** legacy
  reference; the production module never ships two competing Dec128 engines.
- Go 1.27 CI and the performance gate pass before the atomic M2 switch.

## Explicit limitations of this M0 capture

This snapshot records declarations, a selected set of executable behavior
fixtures and baseline results. It does not yet establish exact byte-level
serialization golden vectors, full SQL integration, or upstream equivalence.
Those are OPEN M0/M1 tasks, not passing checks.


## M1 signed-zero and long parser addendum (2026-11-17)

The observable `StoredNegative` state can be true on a **zero with scale 0**,
although the historical `New(0,0,true)` constructor normalizes it. Two
historical paths can produce this state: `Rescale(0)` of a signed-zero value,
and parsing an all-zero long integer with a minus sign. The independent codec
must distinguish this from short `-0`, which normalizes the sign.

Likewise, an already-zero value rescaled to another scale retains its stored
sign, but a nonzero negative number that rounds to a zero coefficient is
normalized by the historical quantizer. Source-level API checks alone cannot
prove these contracts: compare `StoredNegative()` and stored scale, not only
`StringFixed()`. Parser error precedence is length-sensitive; see D19–D22 in
[MIGRATION_DECISIONS](MIGRATION_DECISIONS.md) and the
[M1 checkpoint](../m1-parser-signed-zero-20261117.md).


## M2-B opt-in public package replacement (2026-11-17)

A **whole-package** candidate cutover is now executable under the
`badassitron_m2` Go build tag. The default public `dec128` implementation
continues using historical arithmetic. With the tag, the same nominal public
`Dec128` type delegates coefficient arithmetic to the upstream Jokruger
module; there is no method-by-method public hybrid.

Both functional suites and `go vet` passed locally under Go 1.23.2 and a
strictly external source-snapshot replacement. The tagged public API also
passed both frozen oracles (3,518 + 1,725 records, no unexpected differences)
and separate-client compilation tests. The complete tagged `-race` attempt
did not finish within the sandbox limit: scoped race suites and a selected
Optimus subset passed. The exact module tag, toolchain and full race gates
remain **OPEN**.

The official reproduction command is
`bash tools/migration-m2-cutover/verify.sh` on Go 1.27 with the published
Jokruger v1.4.0 dependency and authentic `go.sum`. The opt-in cutover must
not become the default until those gates and performance exceptions are
reviewed. See the
[M2-B report](../m2b-experimental-public-cutover-20261117.md).
