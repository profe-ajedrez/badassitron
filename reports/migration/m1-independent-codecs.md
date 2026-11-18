# M1 — Independent codec and SQL bridge (17 November 2026)

## Scope

`internal/dec128bridge.Compatible` now parses decimal input, renders text and
JSON, and implements `driver.Valuer` / `sql.Scanner` **without calling the
historical `dec128` parser, formatter or SQL implementation**. Its normal paths
use the imported Jokruger `dec128` and `uint128` module packages. No Jokruger
source was copied or vendored.

The experimental `Legacy()` and `WrapLegacy()` conversion helpers still import
Badassitron's public `dec128` package. They are deliberately retained for
migration interoperability and differential tests. Thus M1 is **not yet fully
independent of the historical implementation**, and the public `dec128` type
has not been switched.

## Behavior and tests

- The parser maintains the historical fixed-point grammar, empty-string zero,
  rejection of exponent syntax, legacy short-input acceptance of `5.`, and
  historical error categories. It uses `ju.FromString` for the unsigned
  coefficient and preserves signed-zero metadata and stored scale.
- Text and JSON output keep the legacy behavior: plain text decimal,
  JSON string values, zero serialized as `0`, and NaN serialized as `NaN`.
- SQL scan accepts `nil`, `string`, `[]byte`, `int`, `int64`, `float32`, and
  `float64`. Unsupported or invalid values leave the receiver unchanged.
- Deterministic differential tests cover 20,000 generated inputs plus 20
  targeted inputs, as well as scan types and rollback behavior. The tests
  compare stored scale, finite sign, result/error class and encoding.
- All repository tests and `go vet ./...` passed; race tests passed for
  `dec128/...` and `internal/dec128bridge`.

## Local benchmark

Host: Intel Xeon Platinum 8573C, Go 1.23.2, linux/amd64.
These measurements are **not comparable to runs with Go 1.27**.

| Operation | Historical | Bridge | Allocations |
|---|---:|---:|---:|
| Parse (long fixed-point input, after stack-buffer change) | ~54.5 ns/op | ~118 ns/op | 0 / 0 |
| String (before stack-buffer change) | ~93–104 ns/op | ~94–103 ns/op | 1 / 1 |

The bridge parser is still about 2.2x slower for this input. It is a
correctness checkpoint, **not** a performance optimization. The benchmark
source is `internal/dec128bridge/compatible_codec_bench_test.go`.

## Verification limitations

Only an uploaded snapshot of upstream `main` was available locally. Tests
used a **temporary local module replacement** and temporarily lowered both
module Go directives to run with the available Go 1.23.2 compiler. The
released archive keeps `go 1.27`, `require github.com/jokruger/dec128 v1.4.0`,
and no local `replace`. The upstream v1.4.0 tag and checksum have not been
verified; the current `go.sum` remains empty. Therefore the archive cannot
be claimed reproducible against the published dependency in this environment.

## Next gates

1. Test against a verified upstream v1.4.0 module and compatible Go toolchain;
   generate and commit the resulting `go.sum`.
2. Replace remaining legacy conversions with a self-contained representation
   and validate all 5,243 frozen oracle observations against a candidate M2.
3. Resolve the currently accepted parser-performance regression and rerun
   benchmarks on the target toolchain.
4. Perform an atomic public `dec128` switch only after all above gates pass.
