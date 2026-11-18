# Frozen public API baseline

The two JSON files were generated from the historical sources inside `badassitron_M3F_imported_default_20261117.zip` on 2026-11-17, before removal of the build-tagged rollback implementation. The `migration-api` source analyzer uses only Go standard-library AST parsing. `historical-public-identifiers.json` covers the 1,030 historical public names (not binary compatibility); `historical-uint128-signatures.json` records the complete 66 unsigned declarations and their source signatures. API runtime and behavioral tests are mandatory in addition to these snapshots.

Reference M3-F ZIP SHA-256: `da3e6c23360470ec9a1040613f14bc33d86f8560dd9602527fbd32f391b14610`. The original ZIP is retained separately by the project owner; historic arithmetic Go source is **not** shipped inside M7-A.
