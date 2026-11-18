# Release readiness (maintainers)

**Status: NOT APPROVED FOR PUBLICATION.** This source snapshot is not
announced as a released version. Verification results for earlier source
checkpoints do not automatically certify the final package.

## Current source

- The root `go.mod` declares **Go 1.27** and imports
  **`github.com/jokruger/dec128 v1.4.0`**.
- `dec128` and `dec128/uint128` delegate 128-bit arithmetic to the imported
  dependency. There is no included alternate 128-bit backend or `vendor` tree.
- The historical API inventory preserves **1,030** public identifiers, plus
  two recorded additive declarations; all **66** historic `uint128` public
  signatures are preserved.
- Deprecated rollback tags are unsupported or inert; they are not alternate
  production implementations.

## Evidence available

- Earlier Go 1.27.1 verification logs and performance studies are preserved
  in the [engineering history](MIGRATION_ARCHIVE.md), but identify their own
  dated source checkpoints.
- The documentation-cleanup source was checked locally with Go **1.23.2**
  using a separately supplied **unversioned Jokruger `main`** snapshot and
  a temporary dependency override outside the deliverable. Tests, race,
  vet, and thirteen example outputs passed under those conditions.
  This is **not** verification against the published v1.4.0 dependency.
- Documentation links and navigation coverage, the public API inventory,
  architecture guard, and packaging hygiene were checked on the final
  documentation source.

## Required before publication

1. Use Go **1.27.1 or newer** on the exact candidate checkout and resolve
   the **officially published Jokruger v1.4.0 module**.
2. Run the full technical verification command:

   ```sh
   GOWORK=off bash tools/release/verify.sh
   ```

   It tests the packages, race behavior, API, independent oracles, module
   identity, documentation links, and runnable example output.
3. Confirm the final source checksum and retain the complete official logs.
4. Complete the **historical source provenance, attribution, and
   redistribution-rights review**. The MIT declaration and existing
   third-party notices do not independently prove all historical authorship.
5. Obtain explicit publication approval only after the technical and
   provenance requirements are satisfied.

No automated test can substitute for the provenance review. The currently
supplied source has **not** passed the final official-host gate.

[Technical verification instructions](../tools/release/README.md) ·
[Provenance ledger](PROVENANCE_REVIEW.md) ·
[Attribution notice](../NOTICE.md) ·
[Engineering records](MIGRATION_ARCHIVE.md)
