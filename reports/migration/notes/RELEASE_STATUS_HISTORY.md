# Release readiness (maintainers)

> **Archived release documentation.** Superseded by the
> [current release checklist](../../../docs/RELEASE_STATUS.md).

**Status:** source prepared for independent review; **NOT APPROVED FOR
PUBLICATION**. No release/tag or external repository change is asserted by
this document. This page records the checks required for the current source;
older migration files are dated historical records.

## Implementation in this checkout

- `go.mod`: `go 1.27`, `require github.com/jokruger/dec128 v1.4.0`.
- `dec128` and `dec128/uint128`: one imported 128-bit arithmetic backend.
  Historical arithmetic Go sources are not part of the production tree.
- Exported declaration inventory: 1,030 historical identifiers preserved,
  with two recorded additive public declarations. All 66 public declarations
  in `dec128/uint128` retain their source signatures.
- Deprecated rollback tags `badassitron_legacy` and
  `badassitron_uint128_legacy` are intentionally unsupported. Tags
  `badassitron_m2` and `badassitron_m3` are inert compatibility aliases.
- The active parser retains the stable Jokruger-delegating implementation.
  Experimental parser candidates were not promoted.

## Different kinds of evidence must not be conflated

| Evidence | What it establishes | What it does not establish |
| --- | --- | --- |
| Supplied M7-D Go 1.27.1 verification | The historically verified **M7-D source** passed its reported official technical checks, including full race and 5,243 frozen cases | That the M7-H package has itself been reverified on an official Go host |
| Supplied M7-D/M7-E official interleaved A/B summary | Reported M7-E `ParseShort` median improvement was **2.1%**, below acceptance | A 3% improvement or permission to promote M7-E |
| M7-G local Go 1.23.2 + supplied Jokruger `main` snapshot | Local adapter build/tests, race, oracles, and diagnostic benchmarks under those explicitly different conditions | Exact Jokruger **v1.4.0** source equivalence or Go 1.27.1 performance |
| Repository hygiene and static validation in M7-H | Source identity, file consistency, public API guard, documentation links, file packaging checks | Complete runtime validation or any legal clearance |

Read the [M7-G report](Informe_M7G_parser_final.md),
[M7-F checkpoint](../m7f-interleaved-evaluation-20261118.md),
and [M7-D report](../m7d-addcarry-review-20261118.md)
for primary context. Do not re-label locally generated results as official.

## M7-H local diagnostic checks (not the official gate)

A disposable copy was tested with **Go 1.23.2** and the supplied unversioned
Jokruger `main` source. Only in that copy, the module directives were lowered
to `go 1.23` and an external local `replace` was added. Results:

- `go test ./... -count=1 -timeout=20m`: PASS.
- `go vet ./...`: PASS.
- `go test -race -p=2 ./... -count=1 -timeout=20m`: PASS.
- Compatibility-tag tests: PASS.
- Both frozen oracles: 3,518 + 1,725 records, no unexpected differences.
- Independent exact-result checks: PASS.
- Eleven runnable examples: PASS.
- Static backend/API/hygiene checks and documentation links: PASS.

Logs, exit statuses, and checksums are retained under
`../reports/migration/m7h-local-validation/`. None is a Go 1.27.1 or an
official Jokruger v1.4.0 result.

## Mandatory gates before publication

1. Run `GOWORK=off bash tools/release/verify.sh` on the **exact final source**
   using Go 1.27.1 or newer and the officially resolved Jokruger v1.4.0 module.
2. Confirm the current `go.sum`, downloaded module identity and official
   LICENSE/NOTICE bytes with the included audit. The attached `main` archive
   does **not** certify the release tag.
3. Complete a file-by-file historical source ownership, attribution and
   redistribution-rights review for Badassitron. Existing LICENSE and third-
   party notices alone are not a provenance chain.
4. Review all warnings and failures, including emitted artifacts, before
   authorizing a tag or release. A passing CI badge is not legal approval.

**Open blocker:** historical source provenance and any necessary permissions.
**Unperformed on this documentation-updated source:** full official-host rerun,
unless subsequent logs explicitly identify the checksum of this exact package.
The additional runnable examples were checked locally with Go 1.23.2 and the
supplied unversioned upstream snapshot. That is diagnostic, not official.

[Provenance ledger](../../../docs/PROVENANCE_REVIEW.md) · [Attribution notice](../../../NOTICE.md) · [Third-party register](../../../THIRD_PARTY_NOTICES.md) ·
[Release tooling](../../../tools/release/README.md)
