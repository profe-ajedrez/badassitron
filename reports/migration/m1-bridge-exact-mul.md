# M1 — Exact multiplication bridge hardening

Date: 2026-11-17. Status: **experimental bridge validated; public engine unchanged**.

## Sources and dependency boundary

- Base: `badassitron_M1_jokruger_bridge.zip` (previous deliverable).
- Upstream test source: user-provided `dec128-main.zip`, snapshot from `main`, archive identifier `367475f9b33250c59b148d382fec07314ca050d3`. **This is not a verified copy of v1.4.0.**
- Production dependency remains `require github.com/jokruger/dec128 v1.4.0` in `go.mod`.
- No upstream source, test implementation, local `replace` directive, or vendored dependency is included in the delivered ZIP.
- Local validation used Go 1.23.2 with *temporary* changes to both modules' `go` directives and a local `replace` pointing at an extracted copy of the upstream archive. Those changes are **not** shipped. Normal module validation against Go 1.27 and verified `v1.4.0` is still required.

## Implemented

`internal/dec128bridge.MulStrict` now mirrors the existing `badassitron/dec128.Mul` contract without calling upstream's globally configured `Mul`:

1. Preserve NaN operand precedence and historical NaN reasons; return canonical zero for zero operands.
2. Fast path: standard-library `math/bits.Mul64` for exact nonzero 64-bit coefficient pairs whose combined scale fits within 19. This path reuses Badassitron's prior optimization and contains no copied Jokruger code.
3. General path: imported `jokruger/dec128/uint128.Uint128.MulCarry` to produce the exact 256-bit product.
4. If the product fits 128 bits and its stored scale is at most 19, preserve that scale; if it exceeds 19, remove *all* trailing product zeros before checking representability.
5. If the product exceeds 128 bits, use imported `QuoRem256By128` and upstream `Pow10Uint128` for exact scale reduction, preserving the historical preference for the largest exact reduction.
6. If unsuccessful, retry with canonical operands, matching historical `Mul`; otherwise return `NaN(Overflow)`.

The previous implementation incorrectly returned a result at scale 19 for `2.0000000000000000000 × 0.0000000000000000005`. Badassitron returns scale **18**, because the mathematical result is `0.000000000000000001`. The improved bridge preserves this exact scale. This is an actual API-contract distinction: numerical equality alone is insufficient.

## Test evidence

Executed on Linux/amd64 Intel Xeon Platinum 8573C with temporary Go 1.23.2 workspace:

- Existing bridge tests: finite round-trip, 3,000 differential small multiplications, global-rounding isolation.
- New `TestMulStrictBoundaryDifferential`: 21 × 21 deterministic boundary inputs plus **30,000** seeded random coefficient/scale/sign cases. Compares NaN, fixed string, stored scale, sign and error details against the legacy implementation. **PASS.**
- New `TestMulStrictFiniteAgainstBigInt`: 5,000 seeded products; every finite result was checked against a separately computed exact `math/big.Int` product at the represented scale. **PASS**.
- New independent check of loss-policy isolation: **PASS**.
- `go test ./... -count=1`: **PASS**.
- `go vet ./...`: **PASS**.
- `go test -race ./internal/dec128bridge -count=1`: **PASS**.

Caveat: differential tests show consistency with the historical implementation, **not** correctness across all values or equivalence with the tagged upstream module. The `math/big` check validates finite arithmetic, but does not independently prove whether all NaN outcomes are necessary.

## Performance checkpoint

Command: `go test ./internal/dec128bridge -run '^$' -bench 'BenchmarkMul' -benchtime=200ms -count=2`, Go 1.23.2, same host, 0 B/op and 0 allocs/op on all rows.

| Operation | Samples (ns/op) | Notes |
| --- | --- | --- |
| `MulStrict` improved | 4.671, 4.962 | Internal bridge; exact old contract |
| `legacy.Mul` | 3.883, 3.319 | Historical fast path |
| upstream `MulRound(..., ROUND_NAN)` | 6.010, 5.590 | Explicit scale |

Previous bridge prototype measured **20.39 ns/op** (median, previous run). This is a material improvement, but the improved bridge is **still slower than** the historical `Mul` fast path in this microbenchmark. These are not general library rankings; rerun under official Go 1.27 and tagged upstream v1.4.0.

## Gating decisions

- **M1 exact multiplication bridge: validated in local untagged snapshot**, without altering published API behavior.
- **M2 atomic replacement: NOT authorized**. Remaining: verified upstream tag/toolchain; method-by-method equivalence for Add/Sub/Div/rounding/parse/serialization/SQL; `uint128` API and error mappings; all 5,243 golden oracle cases against the *candidate replacement*; benchmark thresholds and final provenance audit.
- `dec256`, `dec512`, and `optimus` remain unchanged.
