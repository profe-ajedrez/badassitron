# Documentation cleanup and navigation review

Date: 2026-11-18  
Scope: user documentation, example programs, archival organization, and
non-production verification tooling.

## Decision

**Completed as a documentation-only release-candidate cleanup.** Production
behavior and existing Go APIs are unchanged. The source remains **NOT
APPROVED FOR PUBLICATION** until official module/toolchain verification and
historical source-provenance review are completed.

## User-facing changes

- Replaced the migration-first root README with an overview, a complete
  calculation example, a type-choice table, and task-based links. Its length
  decreased from 167 to 111 lines before the final packaging pass.
- Rebuilt `docs/README.md` as a discoverable index with separate Getting
  Started, Everyday Tasks, API Reference, and Maintainers sections.
- Created `docs/RECIPES.md` (ten common tasks with code, runnable commands,
  expected values, or direct links) and `docs/OPTIMUS.md` (practical adaptive
  decimal guide).
- Created `docs/UNSIGNED_INTEGERS.md` and simplified
  `dec128/PRECISION.md` and `optimus/CONCURRENCY.md` for present-day use.
- Reorganized the example index and added two runnable programs:
  `examples/12_uint128` (checked integer parsing, overflow) and
  `examples/13_rounding_boundaries` (per-line rounding versus rounding a sum).
- Replaced the engineering milestone-led public `CHANGELOG.md` with a concise
  unpublished-source changelog; moved the complete historic engineering
  changelog to `reports/migration/ENGINEERING_CHANGELOG.md`.
- Shortened README, NOTICE, third-party notices, release instructions,
  and release-readiness statements without claiming legal clearance or
  official verification that had not occurred.
- Moved eleven stage-specific documentation files out of the active docs
  path into `reports/migration/notes`, with updated relative links and
  explicit historical-record notices. Archived the original precision,
  concurrency, release-status, and release-tooling notes separately.

The normal user-documentation route contains no engineering milestone labels.
Technical history, benchmark evidence, and original reports remain accessible
through the maintainer-only engineering history index; no audit evidence was
silently reclassified as current product behavior.

## Source and dependency integrity

Compared against the supplied `badassitron_M7H_clean_repository.zip`:

| Check | Finding |
| --- | --- |
| Previously existing `.go` files | **175 / 175 byte-identical** |
| Added `.go` files | Exactly two, both inside `examples/` |
| `go.mod` and `go.sum` | Byte-identical |
| API inventory | 1,030 historical public identifiers preserved; two documented additions remain |
| `uint128` API | All 66 historical signatures preserved |
| Backend architecture | Imported Jokruger v1.4.0 remains the sole production 128-bit arithmetic backend |
| Temporary module changes | None in the delivered repository |
| `vendor`, `replace`, `go.work` | Absent |

`tools/release/verify.sh` was extended with two documentation checks. No
arithmetic implementation, decimal public API, or package module directive
was changed.

## Documentation and example verification

| Check | Result |
| --- | --- |
| Markdown paths | **489 valid local links across 122 Markdown files** before adding this report |
| Navigation | Root entry points, nine main guides, and thirteen runnable examples indexed |
| Python release-tool tests | **28 PASS** |
| Source backend guard, package hygiene, public API guard | **PASS** |
| All 13 example stdout blocks compared exactly to README expected output | **PASS** (diagnostic environment below) |
| `go test ./... -count=1 -timeout=20m` | **PASS** (diagnostic environment) |
| `go vet ./...` | **PASS** (diagnostic environment) |
| `go test -race -p=2 ./... -count=1 -timeout=20m` | **PASS** (diagnostic environment) |
| `GOWORK=off bash tools/release/verify.sh` on available official-source tree | **BLOCKED**: installed Go 1.23.2 is below declared Go 1.27 |

For the three Go verification commands and the example output check, an
isolated disposable copy was used with Go **1.23.2** and a supplied,
**unversioned Jokruger `main` source ZIP**. Only that disposable copy lowered
both modules' Go directives to 1.23 and set a relative local `replace`.
The source ZIP does not establish identity with **official Jokruger v1.4.0**.
Those local PASS results must not be relabeled as official Go 1.27.1 results.
The official verification command rejected this environment with exit 1.

## Reproduce with the official dependency

From the exact candidate checkout, using Go 1.27.1 or later and the official
pinned dependency:

```sh
GOWORK=off python3 tools/check_doc_links.py
GOWORK=off python3 tools/check_doc_navigation.py
GOWORK=off python3 tools/check_example_outputs.py
GOWORK=off bash tools/release/verify.sh
```

The example output checker executes every `examples/*/main.go` program and
compares its stdout byte-for-byte with that folder's fenced `Expected output`
block. The navigation checker confirms that every runnable example is linked
from `examples/README.md` and the main guides are linked from `docs/README.md`.

## Remaining boundaries

- An independent, exact-source Go 1.27.1 + authenticated Jokruger v1.4.0
  release-gate run is still required.
- The previous technical verification evidence applies to its identified
  historical source, not automatically to this new documentation package.
- Historical Badassitron source ownership, licenses, attribution, and
  redistribution permissions have **not** been fully cleared. Retained MIT
  and third-party notices are disclosures, not independent legal approval.
- Link checking verifies files but not all Markdown fragment anchors or
  off-site URLs. The output check covers the included sample inputs, not
  every possible numerical edge case.

No release tag was created and nothing was published.
