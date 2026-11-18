# Source provenance review ledger

**Review scope:** source material in this repository snapshot, not a legal
opinion, ownership determination, or assertion about all earlier revisions.
**Overall status: OPEN — publication approval withheld.**

## Verified from the delivered files

| Surface | Local evidence | Finding | Remaining question |
| --- | --- | --- | --- |
| Root `LICENSE` | MIT text naming the declared Badassitron holder | License declaration exists and was retained without alteration | Independent ownership and permission for every historical source file |
| `go.mod` and `go.sum` | `go 1.27`, exact Jokruger v1.4.0 requirement, recorded module checksums | Pinned imported dependency, no `replace` or vendor source | Rerun resolution and `go mod verify` of v1.4.0 on the final official host |
| Jokruger notices | [retained MIT LICENSE](../third_party_licenses/jokruger-dec128-snapshot/LICENSE) and [NOTICE](../third_party_licenses/jokruger-dec128-snapshot/NOTICE) | Retained hashes match prior supplied audit hashes; the NOTICE includes BSD-3-Clause and MIT contributor attributions | Fresh verification of the **published v1.4.0** bytes and all downstream requirements |
| `dec128`, including `uint128` | Source and import inspection; architecture guard | The former built-in 128-bit engine is absent; the wrapper imports Jokruger arithmetic | Provenance of Badassitron-owned wrapper and remaining decimal/codec code over its history |
| `dec256`, `dec512`, `internal`, `optimus` | Current source files; no provenance metadata proving their creation history | These are Badassitron source files, **not automatically verified as independently authored** | File-by-file historical comparison with authorship, chronology and permissions |
| `withdec128`, `withfloat64`, examples and tooling | Source files and repository MIT declaration | Application handlers and examples are included; no full historical ownership verification | Review legacy contributions, copied fragments and external inspirations where applicable |
| Frozen API/oracle testdata | Test fixtures with baseline metadata, no historical engine code | Useful reproducibility evidence, not proof of independent ownership | Confirm redistribution rights and provenance of any fixture derived from third-party content |
| `reports/` history | Dated experiments and supplied-host reports | Audit evidence, **not** a statement that old rollback options are active | Retain source labels and validate any archival claims before quotation |

## What was *not* established

- No complete commit-by-commit or file-by-file historical ownership audit.
- No independent comparison of all older algorithms with possible upstream
  sources; similarity by itself establishes neither copying nor non-copying.
- No live Go 1.27.1 download/verification of the officially tagged v1.4.0
  dependency **for this source snapshot** in the available container.
- No legal clearance to apply the repository's MIT declaration to code whose
  historical provenance remains uncertain.

## Evidence handling and required follow-up

1. Freeze the candidate source hashes and link the exact final Git tree or
   distributable checksum to all official verification logs.
2. Review the Git history and each historically significant algorithm for
   authorship and license ancestry, including any earlier code no longer in
   production, if it influences current files or notices.
3. Where outside material is confirmed, document its identity, version,
   license, original copyright and any required acknowledgments; replace or
   seek permission for any incompatible material.
4. Repeat module LICENSE/NOTICE identity audit against official v1.4.0;
   preserve the result under the exact final source checksum.
5. Have the rights holder review and explicitly approve this ledger before
   calling the package published or provenance-cleared.

The resulting release gate consists of **independent technical validation
AND provenance clearance**. Neither can substitute for the other.

[Current release status](RELEASE_STATUS.md) · [Attribution statement](../NOTICE.md)
