# Migration record index (historical)

The M0–M7 reports describe dated engineering snapshots and frequently use
words such as *current*, *default*, *pending* and *rollback*. These statements
are true **only within each dated checkpoint**, not instructions for this
M7-H source checkout. Read [current release status](RELEASE_STATUS.md) first.

| Stage | Historical records | Live disposition |
| --- | --- | --- |
| M0–M2 | [Master plan](../reports/migration/notes/MIGRATION_MASTER_PLAN.md), [decision ledger](../reports/migration/notes/MIGRATION_DECISIONS.md), [contract](../reports/migration/notes/DEC128_MIGRATION_CONTRACT.md) | Former whole-package opt-in and rollback mechanisms were retired |
| M3 | [Validation history](../reports/migration/notes/MIGRATION_VALIDATION.md), [performance history](../reports/migration/notes/MIGRATION_PERFORMANCE.md) | Imported `uint128` adapter is default; rollback code removed |
| M7-A | [Source consolidation](../reports/migration/m7a-source-consolidation-20261117.md) | Historical implementation removed from production |
| M7-C | [Adapter tuning](../reports/migration/m7c-adapter-performance-20261118.md) | Historical checkpoint |
| M7-D | [AddCarry optimization](../reports/migration/m7d-addcarry-review-20261118.md) | **Stable production source** retained in M7-H |
| M7-F | [A/B harness](../reports/migration/m7f-interleaved-evaluation-20261118.md) | Experimental tooling and evidence, not a parser promotion |
| M7-G | [Parser decision](../reports/migration/notes/Informe_M7G_parser_final.md) | **Rejected** parser optimization; no production change |

The archived reports are preserved for review; they are not claims of a
published release. Some earlier raw benchmark outputs are retained in a
separate M7-H evidence archive to keep the working-tree ZIP focused. Follow
links within the repository for files kept in-tree.
