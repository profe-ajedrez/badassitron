# M7-D / M7-E official interleaved summary — user-supplied, NOT independently recomputed

Source: figures supplied in the M7-G request dated 2026-11-18. The independently readable
`Archivador(10).zip` and the complete official A/B sample series were NOT attached to
this M7-G execution. Do not attribute bootstrap confidence intervals to these medians.

- Go 1.27.1; 8 ABBA/BAAB blocks; 16 invocations per version; 250 ms per subbenchmark.
- ParseShort: M7-D 6.940 ns/op, M7-E 6.798 ns/op, reported paired change -2.1%; E wins 8/8 blocks.
- ParseShortMixed: M7-D 7.275 ns/op, M7-E 6.973 ns/op, reported change -4.1%; E wins 8/8 blocks.
- ParseLong: M7-D 39.450 ns/op, M7-E 39.545 ns/op, approximately unchanged.
- Decision: M7-E fails the pre-existing >=3% ParseShort threshold; no promotion.

M7-F separately contains historical supplied **non-interleaved** raw official results
under `reports/migration/m7f-supplied-official/` and independent **local** Go 1.23.2
interleaved samples under `reports/migration/m7f-interleaved-local/`. They are not
the missing Go 1.27.1 interleaved M7-D/M7-E raw official archive.
