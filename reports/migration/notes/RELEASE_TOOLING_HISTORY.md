# Technical release verification (maintainers)

> **Archived release documentation.** Superseded by the
> [current release checklist](../../../docs/RELEASE_STATUS.md).

The current code imports **`github.com/jokruger/dec128 v1.4.0`** as the sole
128-bit arithmetic engine. This source tree does not contain its Go source or
the former Badassitron 128-bit arithmetic implementation.

Run from the **exact candidate checkout** on Go 1.27.1 or newer, with the
official Jokruger v1.4.0 module accessible:

```sh
GOWORK=off bash tools/release/verify.sh
BENCHTIME=250ms COUNT=7 bash tools/release/bench-backend.sh reports/migration/m7h-official-backend
```

`verify.sh` fails if the Go version or official module is not available. It
checks pinned module identity and retained upstream LICENSE/NOTICE bytes,
backend architecture, repository hygiene, the historical public API inventory,
the complete test/vet/race suites, both frozen result oracles (3,518 and
1,725 records), independent exact-result checks, compatibility alias tags,
an external consumer module and local documentation links, navigation coverage, and example output checks.

The official license/module audit is written to
`reports/migration/m7h-official-module-audit.md` **only if executed**. No
historical log from M7-D is proof that this M7-H source passed that gate.
`bench-backend.sh` reports same-build timings and allocations for facade
operations versus directly imported Jokruger. These measurements are
workload- and environment-specific; they do not guarantee overall parity.

The independent [release/provenance checklist](../../../docs/RELEASE_STATUS.md)
remains mandatory. **No command here can determine historical authorship or
authorize publication by itself.**

## Build-tag and implementation boundaries

- `badassitron_legacy` and `badassitron_uint128_legacy`: deliberately
  unsupported; they produce compile errors.
- `badassitron_m2` and `badassitron_m3`: inert compatibility aliases, not
  alternate runtime backends.
- `Uint128.Add`: delegates to Jokruger's exported `AddCarry`, retaining the
  historical zero-on-overflow convention.
- `uint128.FromString`: the stable M7-D wrapper is retained after M7-G.
- The full historical source is **not** bundled or re-enabled by build tags.

## Archived measurements and decisions

Historical adapter measurements and parser experiments are preserved for
audit, not as live rollout instructions:

- [M7-C performance](../m7c-adapter-performance-20261118.md)
- [M7-D stable adapter](../m7d-addcarry-review-20261118.md)
- [M7-F A/B methodology](../m7f-interleaved-evaluation-20261118.md)
- [M7-G decision and limitations](Informe_M7G_parser_final.md)
- [Migration record index](../../../docs/MIGRATION_ARCHIVE.md)

To run an **experimental** parser A/B comparison with an independently
provided candidate archive, use `tools/release/bench-interleaved.py --help`.
It does not change production or remove the need for full official-host
validation. Never treat the Jokruger `main` source ZIP as a verified match
for the pinned v1.4.0 tag.

To verify that documented outputs remain correct:

```sh
GOWORK=off python3 tools/check_example_outputs.py
```

This executes all examples; `go test ./examples/...` by itself only compiles them.
