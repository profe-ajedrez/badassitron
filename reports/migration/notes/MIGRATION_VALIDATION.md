> **Historical migration record (archived at M7-H).** This file contains

> **Archived engineering record.** This document describes a dated source
> checkpoint, not the currently recommended API or release state.
> For supported usage, see the [documentation index](../../../docs/README.md).
> time-specific stage status, proposed code and, in places, old rollback
> tag instructions that **do not apply to the current source tree**.
> The authoritative state is [M7-H release status](../../../docs/RELEASE_STATUS.md);
> do not follow an old "current" or "pending" claim as a live rule.

## M3-C unsigned `Mul` correctness and inlining checkpoint

## M3-F current default unsigned backend (unreleased)

The default `dec128/uint128` arithmetic and codecs now delegate to imported
Jokruger v1.4.0. `badassitron_uint128_legacy` selects the old unsigned engine
for differential testing; `badassitron_legacy` also rolls back the decimal
engine. `badassitron_m3` is a compatibility alias, not a new candidate.
The local suite, signed/unsigned and wide consumers, both frozen oracles,
and the exported 66-symbol API checker passed with Go 1.23.2 and an external
temporary source snapshot. The Go 1.27.1 official-module gate **has not**
run on this code. Follow [M3-F report](../m3f-default-imported-uint128-20261117.md).


Six thousand deterministic `Mul` exact/overflow paths are now checked
against `math/big` in both build modes. The M3-tagged full suite, race suite,
client fixture and frozen oracle cases pass locally. The combined
legacy-decimal+M3 race command timed out, but its remaining packages passed
when run separately. Re-run on Go 1.27.1 with official module before release.
See [M3-C report](../m3c-mul-inline-20261117.md).

# Migration validation protocol

## M3-A unsigned adapter (opt-in; experimental)

`-tags badassitron_m3` tests the complete public unsigned arithmetic facade
with imported Jokruger primitives. The historical `uint128` is still the
default. Source/API verification, 12,000 deterministic independent unsigned
operand cases (5,000 + 4,000 + 3,000), complete locally run suites and
race tests, and 5,243 decimal oracle cases pass in this environment. The
[report](../m3-uint128-opt-in-20261117.md) records the
unapproved performance regressions and the Go 1.27.1/official module rerun.

## M2-E current verification policy (unpublished; 2026-11-17)

The candidate is now the default and the historical engine is explicitly
selected using `-tags badassitron_legacy`. `verify.sh` checks both whole-package
source selections, full test/vet/race suites, 5,243 frozen differential cases,
independent exact checks, an external client and official Jokruger checksum
metadata. See the [M2-E report](../m2e-default-promotion-20261117.md).

The supplied **pre-promotion M2-D** Go 1.27.1 verification log is immutable
historical evidence; this environment can rerun only Go 1.23.2 with a temporary
external modfile/snapshot. Do not declare M2-E officially passed or publish a
release based solely on old logs. No local replacement is included in `go.mod`.

## Baseline and toolchain limitations

The source ZIP declares Go 1.27; this execution environment has Go 1.23.2 and
cannot download the Go 1.27 toolchain or Jokruger v1.4.0 from the network.
To run local diagnostics, `go.mod` was **temporarily** changed from `1.27`
to `1.23` and restored before packaging. This is a diagnostic workaround;
**none** of these results constitute an official Go 1.27 pass.

## Reproduce (supported Go 1.27 host)

```sh
GOWORK=off go test ./... -count=1 -timeout=20m
GOWORK=off go test -race ./... -count=1 -timeout=20m
GOWORK=off go vet ./...
GOWORK=off go test ./... -run '^$' -bench=. -benchmem
GOWORK=off GOMAXPROCS=1 go test ./dec128 -run '^$' -bench '^BenchmarkFastPath$' -benchtime=250ms -count=5 -benchmem
GO111MODULE=off go run ./tools/migration-api -root . > /tmp/api-current.json
python3 tools/check_doc_links.py
```

Compare `/tmp/api-current.json` with `testdata/migration/api-baseline.json`.
A difference in source declarations requires review; matching snapshots alone
is not sufficient for binary/source compatibility.

## Differential harness required for M1

1. Keep historical source out of the production module as an immutable
   separately buildable checkout with recorded ZIP SHA-256.
2. Generate input cases in a shared neutral JSON format (input strings,
   method, scale, rounding mode, original error code, expected printed value).
3. Execute the legacy runner and candidate runner in **separate processes**;
   compare value, scale, signed zero, NaN code, error details, serialization.
4. Mark only ledger-approved intentional corrections as divergences; never
   rewrite expected results to make the new engine pass.
5. Add boundary-heavy seeded fuzzing: scale 0/19, max 128-bit coefficient,
   factor powers of ten, cancellation, negative zero, ties, invalid states.
6. For values beyond the Dec128 range, use `math/big.Int`/`math/big.Rat` as
   an independent exact rational reference; do not compare rounded strings
   as a substitute for mathematical equivalence.
7. Verify global-policy independence in isolated processes, avoiding mutation
   of global configuration during parallel tests.
8. Repeat for `dec256`, `dec512`, `optimus` before promotion between tiers.

## Performance method

Record Go version, CPU, OS, flags, GOMAXPROCS, raw output, `ns/op`, `B/op`,
`allocs/op`, and distribution over repeated runs. Comparing this sandbox's Xeon
with the historical Ryzen FASTPATH_V4 table is **not** a controlled speedup or
regression result. Benchmark old, upstream-direct, and new adapter on the same
host at equal scales and rounding/error policies before claiming a ratio.

## M1 executable baseline (new)

See [`tools/migration-oracle/README.md`](../../../tools/migration-oracle/README.md)
and [`reports/migration/m1-oracle.md`](../m1-oracle.md).
The deterministic 3,518-case JSONL baseline is checked into
`testdata/migration/oracle/`. Run the candidate process and strict comparison
before any atomic M2 replacement. Every intentional divergence must be
approved by exact case ID, output field, old value, new value, and decision ID.

## M1 supplemental oracle

See [the M1 extension report](../m1-extension.md). The
second immutable corpus adds `Mod`, `QuoRem`, `PowInt`, and the explicitly
lossy `EncodeToUint128`. Both original and supplemental streams must be
replayed as separate gates. An oracle observation never automatically
legitimizes a historical bug.


## 2026-11-17 continuation results

The new [parser and signed-zero report](../m1-parser-signed-zero-20261117.md)
records reproducible raw output for the experimental bridge tests, two frozen
oracle runs totaling 5,243 cases, exact `math/big` parser checks, all examples,
vet, and scoped race checks. The Go toolchain remains **1.23.2 only for local
temporary testing**, while `go.mod` continues to require Go 1.27. The published
upstream module/tag and checksum were not fetched; the local `main` reference
is not an authenticated substitute for v1.4.0.

The full `go test -race ./...` run timed out in this sandbox. All packages
other than `optimus` passed a separate full-package race run, and a selected
`optimus` race subset passed. These results do **not** clear the M1/M2 full-race
gate. The frozen 5,243-case oracle still tests the legacy public API and is
not evidence that an atomic M2 candidate already exists.

## M1-final / M2 candidate-A — 2026-11-17 (offline/local-only)

An independent `internal/dec128candidate` implementation now replays **both**
existing frozen oracles without calling legacy public arithmetic:
**3,518/3,518** primary and **1,725/1,725** supplemental cases, no
unapproved differences. The exact Decimal checks report 389 primary and
1,634 extended checks with zero mismatches. It also passes randomized
arithmetic/rounding/sqrt tests, explicit signed-zero tests and a serial
upstream-global-policy independence test. Surface parity on exported AST
signatures is recorded, but not assumed sufficient for release.

The local full `go test ./...`, `go vet ./...`, and `go test -race ./...`
runs passed (see the checkpoint report for raw logs). The previous incomplete
race attempt is retained as historical evidence rather than overwritten.

The local toolchain remains **Go 1.23.2** using a **temporary** `replace` to an
unverified upstream-main snapshot outside the module. The *delivered* module
has `go 1.27`, `require github.com/jokruger/dec128 v1.4.0`, and no `replace`.
`go.sum` has not been generated from the official version. Therefore these
passes are **not** official upstream/Go1.27 release certifications.

See [M1-final / M2 candidate-A report](../m1-final-m2-candidate-a-20261117.md)
and [reproduction instructions](../../../tools/migration-candidate-oracle/README.md).


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


## Official M2-B evidence versus M2-C changes — 2026-11-17

The user-supplied [Go 1.27.1 official-module verification log](../m2b-user-supplied-go1271-verify-20261117.txt)
reports both full package suites, both full race suites, external clients,
3,518 + 1,725 cases and independent mathematical checks passing **on M2-B**.
This is evidence from the supplied execution, not a new run in the current
offline container. The later **M2-C** candidate modifications and new
independent `math/big` tests require the same official verification again.
The release `go.sum` is still empty in the supplied M2-B archive.

Run `bash tools/migration-m2-cutover/verify.sh` with Go 1.27.1 and the
published module, then `BENCHTIME=200ms COUNT=5 bash
 tools/migration-m2-cutover/bench-m6.sh <output-dir>` on the same machine.
The release decision must inspect the recorded performance exceptions and
[third-party notices](../../../THIRD_PARTY_NOTICES.md).


## M2-D independent division and official metadata gate

[M2-D report](../m2d-fast-division-provenance-20261117.md): 1,507 independent rational input pairs × six scales × seven rounding modes, plus 5,243 historical frozen cases. Both complete local race suites passed. The official [cutover verification](../../../tools/migration-m2-cutover/README.md) now checks published `h1:` download metadata against `go.sum` and records official MIT license/NOTICE SHA-256 comparisons. This latter step requires a reachable official module and was not run in the local offline validation.

## M3-E coefficient/backend boundary controls

The [M3-E report](../m3e-architecture-resolution-20261117.md)
records the four build configurations, 4,000 exact constructor-boundary cases,
two frozen oracles (3,518 + 1,725) under both default and M3 configurations,
and bounded local race coverage. Full official Go 1.27.1 tests and historical
provenance approval remain release requirements.

On a Go 1.27.1 host with the published v1.4.0 dependency, run
`bash tools/migration-m3-architecture/verify.sh` to enforce the source
boundary, nominal public type, M2 mirror and existing API/build selection
checks. It must supplement rather than replace the full M2 and M3 verification
scripts. The static source guard is intentionally narrower than full type-
checked whole-program analysis.
