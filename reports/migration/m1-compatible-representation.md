# M1 checkpoint — lossless representation and codec adapter

Date: 2026-11-17. Input: `badassitron_M1_direct_division.zip`.

## Scope and implementation

- Added **`dec128.Dec128.StoredNegative()`** as an additive, read-only accessor for the stored sign of finite values, including negative zero. `IsNegative`, `Sign`, and numerical comparisons keep their existing semantics. NaN returns `false`.
- Added `internal/dec128bridge.Compatible`: an **experimental** private-field wrapper holding the **imported** `github.com/jokruger/dec128.Dec128` plus explicit metadata for the stored sign and scale of zero. `WrapLegacy`, `NewCompatible`, and `Legacy` permit lossless representation roundtrips for legacy finite values, including signed zero. The earlier raw `ToUpstream` / `FromUpstream` helpers are still lossy for signed zero and must not be used as a replacement for this wrapper.
- Added compatibility-surface adapters for string, fixed string, text, JSON, SQL `driver.Valuer`, and `sql.Scanner`. Parsing and SQL conversion currently call the existing Badassitron grammar and conversion functions, so **these are not yet independent upstream-powered implementations**. Erroring scans and unmarshals do not mutate the destination. The experimental `Compatible.Scan` uses a pointer receiver; historical `dec128.Dec128.Scan` is unchanged.
- The adapter retains the historical JSON/text normalization of zero. These serializations intentionally *do not encode* the stored zero sign or scale; lossless representation is guaranteed only by the in-memory wrapper (and explicitly supplied constructor metadata), not by a serialized roundtrip.

**Dependency policy:** Jokruger remains a Go module import (`require github.com/jokruger/dec128 v1.4.0`); no upstream source is included in this artifact, no `replace` directive is delivered. The supplied `dec128-main.zip` was extracted outside the project only for temporary validation.

## Tests and results

- All 20 stored zero-scale/sign combinations × 2 signs, including scale zero normalization, covered by both legacy-to-wrapper conversion and raw-compatible construction.
- 20,000 seeded finite values: coefficient, scale, stored sign and error roundtrips: PASS.
- Historical NaN reason mapping, invalid grammar, codec output matching, JSON/text roundtrip and SQL `driver.Valuer`: PASS.
- Pointer SQL scanning with `nil`, `string`, `[]byte`, `int`, `int64`, `float32`, and `float64`: PASS; failed input leaves destination unchanged.
- `go test ./... -count=1`: PASS.
- `go vet ./...`: PASS.
- `go test -race ./internal/dec128bridge ./dec128 -count=1`: PASS.

## Environment and validity limitations

Executed with local Go 1.23.2, using temporary Go directive downgrades and a local `replace` **only in the validation workspace** to the user-supplied Jokruger `main` snapshot. The distributed `go.mod` is restored to `go 1.27` and the real `v1.4.0` module requirement. This is not a verified test of the official v1.4.0 release using Go 1.27. The existing `go.sum` is empty and its definitive checksums must be generated in a supported online/official-module environment.

## Remaining M1/M2 blockers

1. Verify official upstream v1.4.0 and a supported Go toolchain. Obtain real checksums.
2. Stop relying on historical parser, JSON and SQL implementations inside the experimental adapter before cutting over `dec128`. Reproduce their behavior with the final wrapper.
3. Determine which operations must propagate signed-zero metadata, not just constructors; add full operation-level invariants and run the cross-process historical oracle.
4. Test performance across representative operations and operand distributions. The public `dec128` implementation remains unchanged.
5. Continue source/provenance audit and explicitly approve any intentional semantic differences before an atomic M2 cutover.
