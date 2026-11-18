# Third-party notices

This file complements [LICENSE](LICENSE) and [NOTICE](NOTICE.md). It lists
known imported material but **does not certify** the provenance of all
historical Badassitron source files.

## Imported Jokruger module

- Required dependency: [`github.com/jokruger/dec128 v1.4.0`](https://pkg.go.dev/github.com/jokruger/dec128@v1.4.0).
- Retained upstream [MIT LICENSE](third_party_licenses/jokruger-dec128-snapshot/LICENSE),
  attributed to **Yurii Biurher (2026)**; recorded SHA-256:
  `2883372912045f50180696b209468e0067e6ff5dc5d80eb55c5054b398067f1c`.
- Retained upstream [NOTICE](third_party_licenses/jokruger-dec128-snapshot/NOTICE)
  includes credits for **Quang (2025, BSD-3-Clause)** and
  **Luke Champine (2019, MIT)**; recorded SHA-256:
  `232c8d04e00ca0d00d4d7e32875f33625cd93313946ab029c3307455ad123109`.
- These retained-file hashes match a supplied historical module audit.
  A fresh authenticated check of **v1.4.0** is required for the exact final
  source. See [release verification](tools/release/README.md).
- Jokruger is imported as a Go dependency; its `.go` sources are not copied
  into this repository.

## Badassitron source

The repository declares MIT, but a complete historical review of authorship,
permissions and third-party source ancestry remains open. No claim of fully
verified exclusive originality is made. See the [provenance review](docs/PROVENANCE_REVIEW.md)
for what is known and what remains to be checked.
