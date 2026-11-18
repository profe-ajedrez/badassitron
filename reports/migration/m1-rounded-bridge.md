# M1 — Explicit rounded operations bridge (experimental)

## Scope

This checkpoint continues from the M1 division/rescale archive. It adds two
**internal-only** methods to `internal/dec128bridge`:

- `MulRoundWithStatus(a, b, precision, mode)`
- `DivRoundWithStatus(a, b, precision, mode)`

Both preserve the established Badassitron call signatures, explicit per-call
rounding, validation order, contextual error mapping, and the returned `inexact`
flag. `MulRoundWithStatus` uses imported Jokruger `uint128` primitives to inspect
the exact 256-bit product and its discarded remainder, and imported Jokruger
`MulRound` to compute the rounded coefficient. `DivRoundWithStatus` uses
Jokruger `DivRoundInexact` directly. No upstream source was copied into this
repository. Neither adapter changes process-wide Jokruger settings.

**Important:** The public `badassitron/dec128` implementation is **unchanged**.
Do not treat these experiments as M2 or as a released migration.

## Test evidence

- Deterministic differential test: 30,000 pseudo-random operand pairs plus
  12 × 12 × 7 × 8 = 8,064 matrix comparisons. Both multiplication and division
  are checked, including invalid modes/precision, division by zero, error code,
  inexactness, numeric coefficient, sign visible through the API, and scale.
- Independent rational-arithmetic oracle: 4,000 generated operand pairs,
  evaluating representable outputs with `math/big.Rat` and truncation at the
  requested scale.
- Isolation test: explicit rounded results remain unchanged when Jokruger's
  arithmetic rounding configuration is changed.
- Complete repository test suite: PASS (see `m1-rounded-tests.txt`).
- `go vet ./...`: PASS (see `m1-rounded-vet.txt`).
- Bridge race detector: PASS (see `m1-rounded-race.txt`).

The historical type cannot expose the internal sign bit of negative zero through
its public getters. Thus these tests establish *observable* equivalence, not
bitwise equivalence. An adapter needs a separate decision and dedicated tests
for internal negative-zero preservation before M2.

## Performance checkpoint

Benchmarks on Intel Xeon Platinum 8573C, Go 1.23.2, three samples per operation,
200 ms target per sample. A single case per operation is insufficient for a
universal performance claim; one division-bridge sample had noticeable variance.

| Operation | Historical samples (ns/op) | Bridge samples (ns/op) | Allocations |
| --- | --- | --- | --- |
| MulRoundWithStatus | 23.15 / 22.24 / 25.15 | 48.71 / 54.73 / 54.11 | 0 |
| DivRoundWithStatus | 26.03 / 21.79 / 23.44 | 39.47 / 38.95 / 88.51 | 0 |

See `m1-rounded-bench.txt`. The multiplication bridge currently does a 256-bit
remainder calculation *and* calls upstream `MulRound`. The duplicate work is a
concrete optimization target; avoid shipping it on the public path as-is.

## Dependency and toolchain limitations

- Jokruger was provided as `dec128-main.zip` (a **main snapshot**, not an
  independently verified v1.4.0 release archive). Its upstream module path is
  `github.com/jokruger/dec128`.
- For validation **only**, local `go.mod` files were temporarily downgraded to
  `go 1.23` and a temporary `replace` pointed to the unpacked Jokruger ZIP.
- The delivery restores Badassitron's declared `go 1.27` and
  `require github.com/jokruger/dec128 v1.4.0`, **without any `replace` or
  vendored upstream sources**. Its official module checksum is still unverified.
- Tests above therefore are not a CI verification with Go 1.27 and the exact
  published v1.4.0 module.

## Exit gate and remaining work

1. Verify official upstream version and Go 1.27 toolchain in a network-enabled CI.
2. Avoid duplicate 256-bit calculation in rounded multiplication; benchmark
   against the legacy FASTPATH_V4 baseline and upstream directly.
3. Decide/verify negative-zero state propagation and all contextual errors.
4. Extend differential testing to parse, formatting, JSON, SQL, and constructors
   before the atomic public `dec128` migration.
5. Maintain Jokruger as a module dependency rather than copying its source.
