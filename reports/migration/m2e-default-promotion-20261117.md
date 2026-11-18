# M2-E — Default-engine source promotion, rollback and release gates

Date: 2026-11-17. **Status: implemented and locally verified; not a published
release or M7 approval.** The starting authority is the supplied complete
`badassitron_M2D_divround_provenance_20261117.zip`, SHA-256
`d0ec78ae3b71055cd491e538314f01d957144cf6c93180d1bc914dde429c9142`.

## Scope and result

The six project-authored `dec128/m2_*.go` files are now the **default public
implementation** of Badassitron `dec128`. They retain the same import path and
are mutually exclusive with the 12 historical implementation files, all of
which now require `-tags badassitron_legacy`. Existing clients need no new
build tag; the former `badassitron_m2` tag selects the new default as an alias.
The `internal/dec128candidate` reference remains mirrored to protect against
accidental divergence. No Jokruger Go code is copied or vendored.

**Rollback:** use `GOWORK=off go test -tags badassitron_legacy ./...` or compile
an application with `go build -tags badassitron_legacy`. This is a source-level
rollback, not a runtime toggle. Do not use the two engines concurrently in a
single binary. Both file sets are checked by
the historical build-selection and candidate-mirror tools, both retained in
the separately archived M3-F source ZIP, not the M7-A release source.
See [the archived tooling notice](../../tools/migration-m2-cutover/README.md).

The default change is **not yet an announced or published version**. Historical
source remains in this development tree solely as an explicit rollback path.

## Accepted implementation changes

1. Inverted the build constraints atomically: `!badassitron_legacy` for all six
   M2 public files and `badassitron_legacy` for the 12 historical files and
   their two white-box tests. No public import path or existing method is
   deliberately renamed or removed.
2. Updated official module verification, build-selection, external-client and
   benchmark scripts for the **default M2 vs tagged historical** arrangement.
   The old `badassitron_m2` tag is checked as an alias. A new source-inventory
   gate confirms the **1,030 public identifiers** visible in the historical
   variant are preserved by the default build; the two additional methods
   are `Dec128.DivAtPrecision` and `Dec128.ErrorCode`. This identifier gate
   does **not** certify every signature or private representation detail;
   the separate external Go client complements it. The legacy private
   `sqlScanDestination.Scan` receiver was renamed to a different private
   receiver in the candidate; it is not an exported type name.
3. Added an independent `math/big` truncation test for 5,040 mixed-width
   coefficient/scale/sign examples (24 seeded cases for each valid pair of
   scales 0–19), plus 400 signed-zero scale combinations and error-precedence
   cases. Both implementations run the same test without private-field access.
4. Recorded user-supplied official Go Modules `h1:` values in `go.sum`.
   The official audit was performed on the user's Go 1.27.1 host; this sandbox
   **did not** independently download the official v1.4.0 archive.
5. Kept the pre-promotion Go 1.27.1 log, module metadata audit and 20-case
   benchmark raw samples intact under
   [`m2e-supplied-evidence/`](m2e-supplied-evidence/). Updated the current
   README, NOTICE, THIRD_PARTY_NOTICES, changelog, migration documents and CI.

A trial `Rescale` optimization using a direct uint64 quotient was measured
locally but was **reverted** because it did not improve the measured path. No
new unchecked arithmetic routine was accepted merely to satisfy the M6
threshold. `Rescale/Truncate` still uses the imported Jokruger uint128 division.

## Official evidence supplied by the user (for unchanged M2-D code)

The received verification record was generated with **Go 1.27.1 linux/amd64**
and the published `github.com/jokruger/dec128 v1.4.0` module. It reports:

- Official `go mod verify` and checksum verification: PASS.
- Full default historical and tagged M2-D test, `go vet` and `-race` suites:
  PASS, including `dec256`, `dec512`, and `optimus`.
- Frozen differential oracles: **3,518 + 1,725 = 5,243**, zero unexpected
  differences; **389 + 1,634 = 2,023** applicable exact arithmetic results,
  zero mismatches. Omitted cases and 38 lossy encodes are not counted.
- External Go client: PASS in both builds.
- Official module checksum: `h1:8+fbylWdTHVZSi839BOXD4Iou6JpNw1euw6n7xRDrDk=`.
- Official module `go.mod` checksum:
  `h1:XmLxm2Jow2CB5IBBR/VhC4MiAkwBzvP9nZBO9n/1VXI=`.
- The separately supplied metadata audit reports exact official `LICENSE`
  and `NOTICE` SHA-256 matches to the retained snapshot texts.

**Evidence integrity:** original uploaded verification log SHA-256
`55493ac472543481a08fa534042272bf62e6e48bf19ad3d26aef6911bbf161f2`;
benchmark ZIP SHA-256
`81413165b68d7a7c01a44a68c7bee574bdea20619f53cb128713cdf595006f88`;
metadata-containing archive SHA-256
`599214b8871d10b0d98579bd3dc8a0cdbb951275f064c96a54cae236ab47e5c8`.
These reports predate the **M2-E build-tag switch and newly added tests**, and
must not be represented as new Go 1.27.1 verification of M2-E.

## Verification executed on M2-E in this session

The sandbox has **Go 1.23.2, Linux amd64, AMD EPYC 9V74** and cannot resolve
`proxy.golang.org` or fetch the Go 1.27 toolchain. Local checks used an
**external temporary `-modfile`** with a replacement pointing at a separately
supplied, unverified Jokruger source snapshot. That modfile and snapshot
are not included in this deliverable; committed `go.mod` still declares
`go 1.27`, `require github.com/jokruger/dec128 v1.4.0` and **no `replace`**.

| Validation | M2-E local result | Evidence |
|---|---|---|
| Default full `go test ./...` | PASS | [log](m2e-local-validation/default-tests.txt) |
| `-tags badassitron_legacy` full tests | PASS | [log](m2e-local-validation/legacy-tests.txt) |
| `go vet` in both builds | PASS | [default](m2e-local-validation/default-vet.txt), [legacy](m2e-local-validation/legacy-vet.txt) |
| Default full `-race` | PASS | [log](m2e-local-validation/default-race.txt) |
| Legacy full `-race` with `-p=2` | PASS | [log](m2e-local-validation/legacy-race-p2.txt) |
| Legacy initial parallel race run | Incomplete (time limit), superseded by `-p=2` PASS | [partial log](m2e-local-validation/legacy-race.txt) |
| Primary frozen oracle | 3,518/3,518, 0 differences | [compare](m2e-local-validation/primary-compare.txt) |
| Supplemental frozen oracle | 1,725/1,725, 0 differences | [compare](m2e-local-validation/extended-compare.txt) |
| Independent exact arithmetic | 389 + 1,634; 0 mismatches | [primary](m2e-local-validation/primary-exact.txt), [supplemental](m2e-local-validation/extended-exact.txt) |
| External Go client, copied temporary Go 1.23 repo | PASS in default and legacy | [default](m2e-local-validation/external-default.txt), [legacy](m2e-local-validation/external-legacy.txt) |
| New 128-bit truncation + signed zero regression | PASS in both builds | [test source](../../dec128/migration_m2e_rescale_test.go) |
| Build-selection and candidate mirror | PASS, 6 M2 default / 12 historical rollback | [scripts](../../tools/migration-m2-cutover/README.md) |
| Migration Python tests (8 + oracle tools) | PASS | [log](m2e-local-validation/python-tools.txt) |
| Public identifier inventory | 1,030 retained; 2 additions; 0 removed | [log](m2e-local-validation/public-api-check.txt) |

Neither the local `-race` result nor the test oracles are a substitute for a
**fresh official-host rerun** against the final M2-E code and official tag.

## M6: 20-case performance and specific exception

The user-supplied official **Go 1.27.1** M2-D comparison ran five samples per
case on AMD Ryzen 7 7735HS. **19/20** sampled cases were within the provisional
+10% threshold (several markedly faster than legacy). The only review case:

| Operation | Legacy median | New engine median | Difference | Allocation |
|---|---:|---:|---:|---|
| `Rescale/Truncate` | 3.54 ns/op | 3.98 ns/op | **+12.6%, +0.44 ns/op** | 0 → 0 B/op |

The new [local Go 1.23.2 paired 20-case table](m2e-local-bench/comparison.md),
using five 100ms samples, measured `Rescale/Truncate` 3.86 vs 3.91 ns/op
(+1.3%) but flagged `Add/OppositeSign` (+13.2%). The disagreement across
compiler versions/hosts is not proof of a regression fix. **D46 accepts the
observed truncation overhead only for this unpublished source-default
checkpoint, not an M7 release**, and D48 keeps final release approval blocked.
Representative workloads and a rerun on Go 1.27.1 are still necessary.

## Pending gates and stage status

- **M0 provenance:** official `LICENSE`/`NOTICE` comparison supplied and
  checksums recorded, but original Badassitron decimal/uint128 source chain,
  attribution sufficiency and permission for redistribution require a human
  source-level legal/ownership review. This is an unresolved M7 gate.
- **M1:** bridge/codecs and their differential test infrastructure are retained;
  local functional coverage and supplied M2-D official verification pass.
- **M2:** default implementation **promoted in source**, previous complete
  engine remains behind explicit rollback build tag. Official final-tree
  Go 1.27.1 testing still required.
- **M3:** migration of the public `uint128` implementation remains pending;
  no speculative changes to wide primitives were introduced.
- **M4–M5:** `dec256`, `dec512`, `optimus` consumers pass in both local builds;
  no changes to their arithmetic representation.
- **M6:** 20-case paired evidence exists; acceptance of real workloads and
  sampled exceptions is pending.
- **M7:** **BLOCKED**: no release tag, default is changed only in this source
  checkpoint. Fresh official M2-E validation, M6 decision and provenance
  approval are mandatory first.

## Reproduce on the official Go 1.27.1 host

From the root of this **M2-E** source checkout (not M2-D):

```sh
GOWORK=off bash tools/migration-m2-cutover/verify.sh \
  2>&1 | tee m2e-official-verify.log
BENCHTIME=200ms COUNT=5 \
  bash tools/migration-m2-cutover/bench-m6.sh \
  reports/migration/m2e-official-bench
cat reports/migration/m2e-official-bench/comparison.md
```

The verify script first checks the official pinned module, the recorded
`go.sum` and actual official `LICENSE`/`NOTICE` bytes. It then validates both
variants, the frozen oracles, exact checks and an external independent module.
If any gate fails, **do not release**: use `-tags badassitron_legacy` as the
rollback for dependent applications while the incompatibility is investigated.
