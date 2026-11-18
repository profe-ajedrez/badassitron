# Badassitron M7-H — Repository Hygiene and Documentation Integrity Review

> **Archived engineering record.** This document describes a dated source
> checkpoint, not the currently recommended API or release state.
> For supported usage, see the [documentation index](../../../docs/README.md).

**Date:** 2026-11-18  
**Scope:** source/package hygiene, current documentation truthfulness, provenance
and release readiness. **No arithmetic or public API changes.**  
**Disposition:** **M7-H hygiene completed; publication approval NOT granted.**
The **M7-D production backend** remains the stable implementation. No
M7-G parser optimization is promoted.

## 1. Inputs and boundaries

Primary input: `badassitron_M7G_parser_final.zip`, SHA-256
`0e3696fd949b5f993695a85df662d9f36c1569174a1eebf4860e20c55ba55b29`.
The supplemental Jokruger `dec128-main(2).zip` is an unversioned snapshot,
SHA-256 `c93593087448a3d888e887fbfcbc6a98abfb9d9ae8484a1e7bb44c38d9101d3b`.
Neither its branch name nor its ZIP hash authenticates **v1.4.0**.

M7-H does not attempt another parser candidate, change divisions, change any
public Go declaration, update the upstream version, or publish a GitHub
release. The work intentionally does **not** interpret an MIT LICENSE file as
proof of a complete historical authorship chain.

## 2. Problems found and corrections

| Finding in M7-G source | Why it was misleading | M7-H correction |
| --- | --- | --- |
| README still described M7-F and stated official M7-D/M7-E interleaved results were pending | The user had supplied the later Go 1.27.1 medians, while M7-G had already closed the parser experiment | Current README explicitly records the 2.1% short-parser improvement, rejection of M7-E, and limitations of the later Go 1.23.2 experiment |
| `NOTICE.md` described the tree as M7-A and deferred its gate | The label was inconsistent with the current stable M7-D backend and M7-G decision | Rewrote the notice around current imported dependency, historical ownership uncertainty and M7-H publication blockers |
| `tools/release/README.md` described M7-D verification as still pending | Historical M7-D technical verification and the current M7-H gate have different source identities | Documentation and CI now identify the new exact-source M7-H technical rerun as pending; M7-D supplied results remain historical |
| Old migration documents used *current*, *default*, and legacy rollback instructions | Their time-specific configurations no longer exist in M7-D/M7-H | Added prominent historical-record warnings to five migration documents, a history index, and one authoritative current status page |
| `docs/FASTPATH_V4.md` and `dec128/PRECISION.md` could suggest the older engine remained active | FASTPATH V4 measurements do not describe the present Jokruger-backed facade | Marked V4 as historical and corrected the active implementation wording |
| No easily discoverable file-by-file provenance review state | A root MIT declaration and upstream NOTICE do not establish historical ownership | Added an explicit provenance ledger, tracked open questions, and retained upstream license texts and their supplied-audit hashes |
| Raw migration logs, a nested model ZIP and Python bytecode were bundled with production source | Obscured a clean release tree, while losing logs would impair auditing | Removed caches; moved unreferenced historical raw evidence to a separately checksummed archive; retained linked records and recent M7-G reproduction evidence |

The main README now separates installation from installing **this specific
unpublished checkout** and distinguishes numerical facilities from any
jurisdiction-specific accounting/tax compliance engine. The changelog retains
historical checkpoint claims as dated history rather than silently rewriting
old test outcomes as if run on M7-H.

## 3. Source and API integrity

Comparison against the original M7-G archive (SHA-256 of each file):

- **All 175 Go files** (including tests, examples and tools) are byte-for-byte
  identical. All **81 non-test Go files** are therefore unchanged.
- `go.mod` is byte-identical: `go 1.27` and
  `require github.com/jokruger/dec128 v1.4.0`.
- `go.sum` is byte-identical; no `replace`, `vendor`, `go.work`, upstream Go
  source copies or additional Go dependencies are in the clean artifact.
- The public compatibility guard reports **1,030 historical public names**,
  **two already-recorded additive declarations**, and **66 exact public
  `dec128/uint128` signatures** unchanged.
- The parser `dec128/uint128/from_upstream_string.go` is unchanged. Production
  still delegates arithmetic to the single imported Jokruger backend.

Updated files are documentation, release-check tooling, `.gitignore` and a CI
workflow; added files include release/provenance/history documents, a
hygiene check and its unit tests, and local validation evidence.

## 4. Repository hygiene and evidence retention

- The M7-G source archive had **766 files**, including 12 Python bytecode
  cache files; the clean M7-H source tree omits those caches.
- **186 older unreferenced raw/evaluation artifacts** were moved, rather than
  discarded, to `badassitron_M7H_historical_evidence.zip` (188 ZIP entries:
  index, manifest and 186 original-path files).
- Files needed by Markdown links, active tests, the release check, and the
  recent M7-G local A/B record remain bundled in the source ZIP. When an
  archival Markdown file links to a directory, that directory was retained.
- A new `tools/release/check_hygiene.py` fails on temporary Python caches,
  nested experiment ZIPs, `vendor`, workspace paths, module replacement and
  missing release/provenance disclosures. Three unit tests cover its positive
  and negative cases.
- `.gitignore` now excludes Python tooling caches and common local build
  scratch. The clean source ZIP has no embedded Jokruger ZIP/source.

The historical evidence archive is **not required to build or test** the
module; it exists for auditability. It has its own `ARCHIVE_INDEX.md` and
`SHA256SUMS`. Historical benchmark records are not current release tests.

## 5. Functional and static results

| Check | Result | Scope |
| --- | --- | --- |
| `python3 tools/release/check_backend.py` | PASS | Current clean source; one imported backend; pinned module |
| `python3 tools/release/check_api.py` | PASS | 1,030 historical names and 66 unsigned signatures |
| `python3 tools/release/check_hygiene.py` | PASS | Clean source layout and current disclosure files |
| Python release unit suite | **28 PASS** | Includes 3 new hygiene unit tests |
| `python3 tools/check_doc_links.py` | **PASS**, 402 local links in the final source | Documentation and retained history links; checks paths, not anchor targets |
| `go test ./... -count=1 -timeout=20m` | PASS **local diagnostic** | Disposable Go 1.23.2 plus supplied Jokruger `main` snapshot |
| `go vet ./...` | PASS **local diagnostic** | Same disposable environment |
| `go test -race -p=2 ./... -count=1 -timeout=20m` | PASS **local diagnostic** | Same; race detection covers exercised tests only |
| Compatibility alias-tag test | PASS **local diagnostic** | Same |
| Primary and extended frozen oracles | PASS, **3,518 + 1,725**, 0 unexpected differences | Same; independent exact checks also PASS |
| Eleven `go run ./examples/<directory>` executions | **11 PASS** | Same; validates runnable examples, not every prose claim |
| `GOWORK=off bash tools/release/verify.sh` on final Go 1.27.1 / official v1.4.0 | **NOT EXECUTED / BLOCKED** | Installed Go is 1.23.2; attempted toolchain download failed |
| Complete historical source licensing/ownership clearance | **OPEN** | Cannot be certified by functional tests or notices alone |

**Critical environment caveat:** the Go functional results above were
produced from an isolated **temporary** clone whose `go.mod` and the provided
Jokruger snapshot's `go.mod` were temporarily set to Go 1.23, using a local
`replace` *only in the disposable clone*. The delivered M7-H module was **not
changed**. The original official-source gate rejected Go 1.23.2 (exit 1),
and automatic Go 1.27 toolchain download was blocked by networking (exit 1).
Neither is presented as a passing official check. Raw logs and exit files
are in `reports/migration/m7h-local-validation/`, with exact commands and
checksums in that directory's README.

## 6. License/provenance findings and limitations

- Repository `LICENSE`: existing MIT declaration retained unchanged.
- Upstream Jokruger LICENSE/NOTICE retained with observed SHA-256 values:
  `2883372912045f50180696b209468e0067e6ff5dc5d80eb55c5054b398067f1c`
  and `232c8d04e00ca0d00d4d7e32875f33625cd93313946ab029c3307455ad123109`.
  The latter records contributor notices attributed to **Quang (BSD-3-Clause)**
  and **Luke Champine (MIT)**.
- The original supplied Go 1.27.1 audit said the retained texts matched the
  published v1.4.0 module; M7-H only independently rechecked local retained
  bytes and did **not** authenticate the tagged module from the network.
- Historical source ownership, code ancestry and any permission/attribution
  requirements beyond the notices remain unresolved. See
  `docs/PROVENANCE_REVIEW.md`, `NOTICE.md` and `THIRD_PARTY_NOTICES.md`.

**Do not release merely because the module declares MIT or the technical
suite passes.** Resolve source provenance separately and document the
rights holder's approval.

## 7. Reproduction and remaining gates

After extracting the full **M7-H source ZIP** on a supported host with
network or trusted official module cache:

```sh
cd badassitron-master
GOWORK=off python3 tools/release/check_backend.py
GOWORK=off python3 tools/release/check_api.py
GOWORK=off python3 tools/release/check_hygiene.py
python3 tools/check_doc_links.py
GOWORK=off bash tools/release/verify.sh
BENCHTIME=250ms COUNT=7 bash tools/release/bench-backend.sh reports/migration/m7h-official-backend
```

`verify.sh` must finish successfully using Go 1.27.1+ and the **official**
Jokruger v1.4.0 module on the exact candidate with recorded hash. It writes
`reports/migration/m7h-official-module-audit.md` after a successful module
resolution and license comparison. Any generated output is new evidence,
**not** part of the existing archive. Before publication:

1. Finish official technical verification and inspect all failures/warnings.
2. Resolve historical source ownership and license provenance and capture
   necessary approvals/notices.
3. Decide whether the desired publication artifact includes the optional
   audit archive; it must not be imported as production Go code.
4. Recheck final module hashes, packaging integrity and tag/source identity.
5. Only then approve a release or tag via a separate deliberate action.

## 8. Final decision

**M7-H repository/documentation cleanup: COMPLETE (technically packaged).**
**M7-H public release readiness: BLOCKED** pending independent historical
source provenance clearance and a fresh official Go 1.27.1 / Jokruger v1.4.0
verification of this exact package. **M7-D remains the stable production
implementation.** No code was published, no release tag was created, and no
performance optimization was promoted by this task.
