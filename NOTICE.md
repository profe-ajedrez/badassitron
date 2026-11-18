# Attribution and provenance

Badassitron contains an MIT license declaration in [LICENSE](LICENSE).
**A complete independent review of historical source ownership, attribution,
and redistribution permissions has not been completed.** The declaration
alone does not establish that every earlier algorithm is independently owned
or cleared for redistribution. Publication approval remains withheld.

## Imported dependency

The source imports `github.com/jokruger/dec128 v1.4.0` as its 128-bit
arithmetic backend. No upstream Go source or vendor directory is bundled.
Retained copies of the upstream LICENSE and NOTICE are under
[third_party_licenses](third_party_licenses/jokruger-dec128-snapshot/), with
attribution and known hashes described in
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

The retained notice hashes were compared with a supplied earlier module
audit. A fresh check against the **official v1.4.0 module** is still required
for this source snapshot. A separately provided, unversioned source archive
is not proof of matching the official tag.

## Remaining review

Historically developed decimal, wide-integer, adaptive-number and handler
code still requires file-by-file authorship and license review. Similarity or
difference to third-party source alone does not establish ownership or
permission. The project's technical tests do not grant legal clearance.

[Current release requirements](docs/RELEASE_STATUS.md) ·
[Maintainer provenance ledger](docs/PROVENANCE_REVIEW.md) ·
[Third-party notices](THIRD_PARTY_NOTICES.md)
