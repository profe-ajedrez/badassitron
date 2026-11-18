# M1 — Exact addition/subtraction bridge checkpoint

Date: 2026-11-17. **Experimental internal implementation; public API unchanged.**

## Dependency and source provenance

The bridge imports `github.com/jokruger/dec128` and its `uint128` package using the module path declared by Badassitron (`v1.4.0`). No Jokruger source code is included in the delivered repository. The user-supplied `dec128-main.zip` was used **only** as the local source of the imported module in a temporary, external test workspace. It is a `main` snapshot, **not verified to match the official `v1.4.0` tag**. No local `replace` is shipped.

The local host uses Go 1.23.2, so test executions required temporarily changing both modules' `go` directives and introducing an absolute-path `replace` in a working copy. The delivered `go.mod` is restored to `go 1.27` and `require github.com/jokruger/dec128 v1.4.0`. `go.sum` remains empty; verification against the actual tag from a network-connected Go 1.27 environment remains mandatory.

## Implemented

- `internal/dec128bridge.AddStrict` and `SubStrict` call Jokruger's explicit `AddRound`/`SubRound` using `ROUND_NAN` with `max(operand stored scales)`. They do not consult upstream's process-wide arithmetic rounding policy.
- The historical behavior of **operand alignment** is preserved: before using the wider exact result from Jokruger, each operand must be individually representable at the target scale. If not, the adapter attempts the legacy canonical-operand retry. This difference was exposed by random differential tests.
- A fast path uses **imported** Jokruger `uint128.AddCarry` / `SubBorrow` for equal-scale same-sign operands, avoiding the generic adapter for common cases. It keeps full-width arithmetic without copying upstream algorithms.
- After an overflow of the primary path, retry canonicalized operands; otherwise return legacy `NaN(Overflow)`.
- The public `badassitron/dec128` methods are **not** rewired yet. `dec256`, `dec512`, `optimus` and compatibility `uint128` are unchanged.

## Validation

Tested in the temporary local workspace (Go 1.23.2, untagged upstream snapshot):

- `TestAddSubStrictDifferential`: 121 seeded boundary pairs and 40,000 seeded random pairs, each checking **both** Add and Sub (80,242 operation comparisons): coefficients, scales, signs, NaN status and error details. **PASS** for the tested set.
- `TestAddSubStrictBigRat`: 12,000 independently generated pairs checked with exact `math/big.Rat` for every finite Add/Sub result (up to 24,000 finite-output checks). **PASS**. NaN representability is not independently adjudicated by this test.
- `TestAddSubStrictIndependentOfUpstreamGlobals`: **PASS**, but does not establish that upstream's mutable setters are concurrency-safe.
- Existing MulStrict, conversion and boundary tests: **PASS**.
- `go test ./... -count=1`: **PASS**.
- `go vet ./...`: **PASS**.
- `go test -race ./internal/dec128bridge -count=1`: **PASS**.

**Known uncovered case:** legacy can preserve the sign bit of zero internally in a way that `IsNegative()` does not expose. Some signed-zero Add/Sub cases therefore cannot be completely reconstructed from the *current* public getters by the experimental adapter. The differential seed set excludes negative zero pending an internal representation strategy. No claim of complete semantic equivalence is made.

## Same-host performance

Intel Xeon Platinum 8573C, Go 1.23.2, 300 ms per run, two samples after the imported-`uint128` fast path (ns/op, zero allocations):

| Microbenchmark | Run 1 | Run 2 | Interpretation |
| --- | ---: | ---: | --- |
| AddStrict | 6.757 | 6.758 | adapter fast path |
| legacy Add | 2.908 | 2.855 | historical fast path |
| SubStrict | 6.767 | 6.648 | adapter fast path |
| legacy Sub | 3.389 | 3.467 | historical fast path |

Before the fast path, AddStrict measured 31.64–32.19 ns/op and SubStrict 35.28–37.10 ns/op in three measurements. The fast path is a significant improvement but **remains slower** than legacy on this specific data. No regression threshold has been approved; do not activate in the published API merely because differential tests pass.

## Remaining gates

1. Independently validate *tagged* Jokruger v1.4.0 and actual Go 1.27 with normal module fetching and checksums.
2. Solve signed-zero representation and ensure the complete API contract (including error state) is satisfiable without duplicating upstream internals.
3. Complete Division, rounding, parse, serialization and SQL compatibility.
4. Run the 5,243-case golden-oracle suite against a separate atomically migrated candidate, and independently check corner cases.
5. Rebenchmark and approve explicit performance regression bounds before enabling M2.
