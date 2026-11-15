# BADASSITRON-POW-OPT-01 — Variable-width multiplication for Dec1024

**Status:** implemented and validated as an experimental optimization; not a production migration decision.

## Objective

Recover the CPU performance of `Dec1024.PowInt` without reintroducing `math/big`, heap allocations, or a new public API. This experiment begins from the unmodified `BADASSITRON-DEC1024-FIXED-01` repository ZIP, not from another branch.

## Source-level findings

`fixedPow` already uses exponentiation by squaring and performs correctly. Its inner `Dec1024.Mul` calls `Uint1024.MulCarry`, which previously multiplied **16 × 16 64-bit limbs** even when coefficients occupied only one or two limbs. In addition, `uint1024.fromWords` reconstructed results via 128 temporary bytes and nested byte decoding. The new implementation:

1. Trims only the unused high limbs of each operand. Leaves the coefficient and full return width unchanged.
2. Calls the existing exact `uintwide.MulFull` over the corresponding significant-length slices; a local 32-word result buffer remains fixed-size, with untouched high words zero.
3. Reconstructs the `Uint1024` value directly from words using nested comparable value structs (`uint128` / `uint256` / `uint512`), rather than intermediate byte arrays.
4. Clarifies the preconditions of `uintwide.MulFull`, which already supports operand slices of differing lengths.

**No change** to public signatures, coefficient width, round modes, scale, error representation, `fixedPow`, or other decimal backends. The core remains free of `math/big`; that package is used in independent test oracles and explicit integer-conversion APIs.

## Verification

Environment: Go **1.23.2**, linux/amd64, Intel Xeon Platinum 8573C. The original `go.mod` declares **Go 1.27**; the tests used separate temporary copies with only the module version directive adapted to 1.23. **Go 1.27 compatibility is not yet verified.**

- `go test ./...`: PASS.
- `go vet ./...`: PASS.
- `go test -race ./dec1024 ./dec1024/uint1024 ./internal/uintwide`: PASS. Full repository race suite was **not** run.
- `TestMulCarrySignificantLimbsAgainstBigInt`: PASS, 17 × 17 limb-width combinations (including zero) and 5 patterns per combination, full 2048-bit product and `Mul` overflow contracts checked against `math/big`.
- `TestPowIntVariableWidthDifferential`: PASS, 12 decimal bases × 17 exponents (positive, negative, zero, `math.MinInt`) compared with `decimalwide.PowInt`, including error and output scale.
- `TestMulCarryNoHeap` and `TestPowIntNoHeap`: PASS with `testing.AllocsPerRun`, no allocations for tested small, mixed and fully-wide integer products, and `PowInt(40)`.
- Existing randomized and differential tests also passed.

## Results

Data are medians of **four alternating baseline/candidate benchmark passes**, with `GOMAXPROCS=1`, `-benchtime=140ms`, `-benchmem`, identical Go version and host. The isolated 1×1 and 16×16 runs add **five paired passes** with `-benchtime=300ms`. Full raw logs and CSV are in the evidence archive. CPU scheduling noise is visible, particularly in the fourth pass.

| Scenario | Baseline | New | Change | allocs/op |
|---|---:|---:|---:|---:|
| `PowInt(40)` on `1.0101` | 7563 ns | 2606.5 ns | **−65.5%** | 0 → 0 |
| `PowInt(-40)` on `1.0101` | 13219 ns | 7735.5 ns | **−41.5%** | 0 → 0 |
| `PowInt(16)` on `1.25` | 5373 ns | 1835.5 ns | **−65.8%** | 0 → 0 |
| `MulCarry` 1×1 | 758.8 ns | 187.4 ns | **−75.3%** | 0 → 0 |
| `MulCarry` 8×8 | 779.2 ns | 391.9 ns | **−49.7%** | 0 → 0 |
| `MulCarry` 16×16 | 894.9 ns | 975.15 ns | **+9.0% in interleaved series** | 0 → 0 |
| `MulCarry` 16×16, isolated recheck | 924.8 ns | 877.5 ns | **−5.1%** | 0 → 0 |

**Interpretation:** substantially faster small-coefficient operations and `PowInt` with zero heap allocations. The 16×16 CPU comparison is **inconclusive** due to sample variation; neither a reliable speedup nor a reliable regression has been established. No benchmark result here is a proof of end-to-end speedup in Iacon or EUCLIDS.

The previously reported historical `math/big` PowInt time is not compared numerically with these runs because it was measured under a different experimental series.

## Risks and next steps

- Verify under the declared Go 1.27 toolchain once available, including escape analysis and full `-race`.
- Run repeated wide-operand performance benchmarks on a quiet/pinned host. Consider a special-case fast path only if a wide regression is reproducible.
- Continue optimizing `MulRound`, `Sqrt`, and other wide arithmetic independently; do not conflate their gains with this change.
- Compare with real Iacon workloads before any production adoption.

**Decision:** retain the significantly improved variant as the next experimental baseline while preserving the previous fixed core and the unmodified upstream repositories as independent references.
